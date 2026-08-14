"""VoxTrace polling worker. Use PIPELINE_BACKEND=whisper for GPU inference."""
import gc, json, logging, os, signal, subprocess, tempfile, time, traceback, uuid
from pathlib import Path
import psycopg
from psycopg.rows import dict_row

logging.basicConfig(level=os.getenv("LOG_LEVEL","INFO"),format="%(asctime)s %(levelname)s %(message)s")
STOP=False
def stop(*_):
    global STOP; STOP=True
signal.signal(signal.SIGTERM,stop); signal.signal(signal.SIGINT,stop)

def claim(conn):
    with conn.transaction():
        row=conn.execute("""SELECT j.id,r.stored_path,r.original_filename FROM jobs j JOIN recordings r ON r.id=j.recording_id
          WHERE j.status='queued' ORDER BY j.created_at FOR UPDATE OF j SKIP LOCKED LIMIT 1""").fetchone()
        if row: conn.execute("UPDATE jobs SET status='processing',stage='preprocessing',progress=5,error_message=NULL,started_at=now() WHERE id=%s",(row["id"],))
        return row

def recover_interrupted(conn):
    changed=conn.execute("""UPDATE jobs SET status='queued',stage='waiting',progress=0,
      error_message='Recovered after worker restart',started_at=NULL
      WHERE status='processing'""").rowcount
    conn.commit()
    if changed: logging.warning("recovered %s interrupted job(s)",changed)

def mock_pipeline(path):
    stem=Path(path).stem
    segments=[
      {"speaker":"SPEAKER_00","start":0.0,"end":6.4,"text":"Rekaman berhasil diterima dan diproses oleh pipeline VoxTrace."},
      {"speaker":"SPEAKER_01","start":6.4,"end":14.8,"text":"Mode demo aktif. Aktifkan backend Whisper untuk transkripsi audio sebenarnya."},
      {"speaker":"SPEAKER_00","start":14.8,"end":21.0,"text":f"Identitas internal rekaman ini adalah {stem}."},
    ]
    return {"language":"id","duration":21.0,"segments":segments,"metadata":{"backend":"mock","pipeline_version":"0.1.0"}}

def is_cancelled(conn, jid):
    row=conn.execute("SELECT status FROM jobs WHERE id=%s",(jid,)).fetchone()
    return not row or row["status"]=="cancelled"

def whisper_pipeline(path, progress=None):
    from faster_whisper import WhisperModel
    device=os.getenv("DEVICE","cuda")
    compute=os.getenv("COMPUTE_TYPE","int8")
    model_name=os.getenv("WHISPER_MODEL","medium")
    model=WhisperModel(model_name,device=device,compute_type=compute)
    language=os.getenv("WHISPER_LANGUAGE") or None
    segments=[]
    duration=float(subprocess.check_output(["ffprobe","-v","error","-show_entries","format=duration","-of","default=nw=1:nk=1",path],text=True).strip())
    chunk_seconds=float(os.getenv("CHUNK_SECONDS","600")); overlap=float(os.getenv("CHUNK_OVERLAP_SECONDS","2"))
    detected_language=language; language_probability=None
    with tempfile.TemporaryDirectory(prefix="voxtrace-") as temp_dir:
        base=0.0; chunk_index=0
        while base<duration:
            extract_start=max(0.0,base-overlap); extract_end=min(duration,base+chunk_seconds+overlap)
            chunk_path=os.path.join(temp_dir,f"chunk-{chunk_index:04d}.wav")
            subprocess.run(["ffmpeg","-loglevel","error","-y","-ss",str(extract_start),"-t",str(extract_end-extract_start),"-i",path,"-ac","1","-ar","16000","-c:a","pcm_s16le",chunk_path],check=True)
            stream,info=model.transcribe(chunk_path,language=detected_language,beam_size=5,vad_filter=True,word_timestamps=True)
            if detected_language is None: detected_language=info.language; language_probability=info.language_probability
            for segment in stream:
                global_start=extract_start+segment.start; global_end=extract_start+segment.end; midpoint=(global_start+global_end)/2
                if midpoint<base or (base+chunk_seconds<duration and midpoint>=base+chunk_seconds): continue
                words=[{"word":w.word,"start":extract_start+w.start,"end":extract_start+w.end,"probability":w.probability} for w in (segment.words or [])]
                segments.append({"speaker":"SPEAKER_00","start":global_start,"end":global_end,"text":segment.text.strip(),"words":words})
                if progress: progress("transcribing",min(85,20+int(65*global_end/max(duration,1))))
            base+=chunk_seconds; chunk_index+=1
    del model; gc.collect()
    return {"language":detected_language or "unknown","duration":segments[-1]["end"] if segments else 0,"segments":segments,"metadata":{"backend":"whisper","model":model_name,"device":device,"compute_type":compute,"language_probability":language_probability,"chunk_seconds":chunk_seconds,"chunk_overlap_seconds":overlap}}

def process(conn,job):
    jid=job["id"]; started=time.monotonic()
    try:
        def progress(stage,value):
            conn.execute("UPDATE jobs SET stage=%s,progress=%s WHERE id=%s AND status='processing'",(stage,value,jid)); conn.commit()
        conn.execute("UPDATE jobs SET stage='transcribing',progress=20 WHERE id=%s",(jid,)); conn.commit()
        result=whisper_pipeline(job["stored_path"],progress) if os.getenv("PIPELINE_BACKEND","mock")=="whisper" else mock_pipeline(job["stored_path"])
        if is_cancelled(conn,jid): logging.info("cancelled job %s",jid); return
        conn.execute("UPDATE jobs SET stage='finalizing',progress=90 WHERE id=%s",(jid,)); conn.commit()
        full=" ".join(s["text"] for s in result["segments"]); result["metadata"]["processing_seconds"]=round(time.monotonic()-started,3)
        with conn.transaction():
            conn.execute("INSERT INTO transcripts(id,job_id,language,duration_seconds,raw_text) VALUES(%s,%s,%s,%s,%s)",(jid,jid,result["language"],result["duration"],full))
            for i,s in enumerate(result["segments"]): conn.execute("INSERT INTO segments(id,transcript_id,sequence,speaker,start_time,end_time,text,words) VALUES(%s,%s,%s,%s,%s,%s,%s,%s)",(uuid.uuid4(),jid,i,s["speaker"],s["start"],s["end"],s["text"],json.dumps(s.get("words",[]))))
            conn.execute("INSERT INTO processing_metadata(job_id,metadata) VALUES(%s,%s)",(jid,json.dumps(result["metadata"])))
            conn.execute("UPDATE jobs SET status='completed',stage='done',progress=100,completed_at=now() WHERE id=%s",(jid,))
        logging.info("completed job %s",jid)
    except Exception as exc:
        conn.rollback(); logging.error("job %s failed: %s\n%s",jid,exc,traceback.format_exc())
        if is_cancelled(conn,jid): return
        conn.execute("UPDATE jobs SET status='failed',stage='failed',error_message=%s WHERE id=%s",(str(exc)[:1000],jid));conn.commit()

def main():
    dsn=os.getenv("DATABASE_URL","postgres://voxtrace:voxtrace@localhost:5432/voxtrace")
    recovered=False
    while not STOP:
        try:
            with psycopg.connect(dsn,row_factory=dict_row) as conn:
                if not recovered: recover_interrupted(conn); recovered=True
                while not STOP:
                    job=claim(conn)
                    if job: process(conn,job)
                    else: time.sleep(float(os.getenv("POLL_INTERVAL","2")))
        except psycopg.OperationalError as e: logging.warning("database unavailable: %s",e);time.sleep(3)
if __name__=="__main__":main()
