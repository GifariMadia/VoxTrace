"use client";

import { ChangeEvent, DragEvent, useEffect, useMemo, useRef, useState } from "react";
import Link from "next/link";

type Status = "queued" | "processing" | "completed" | "failed";
type Segment = { id: string; speaker: string; start: number; end: number; text: string };
type Job = { id: string; filename: string; status: Status; stage: string; progress: number; createdAt: string; createdLabel?: string; duration?: number; segments?: Segment[] };

const API = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";
const seed: Job[] = [
  { id: "demo-001", filename: "observasi-engineering.m4a", status: "completed", stage: "done", progress: 100, createdAt: "2026-08-14T00:30:00.000Z", createdLabel: "35m lalu", duration: 184, segments: [
    { id:"s1", speaker:"Interviewer", start:0, end:8, text:"Bisa diceritakan bagaimana alur deployment tim saat ini?" },
    { id:"s2", speaker:"Engineer", start:8, end:24, text:"Perubahan masuk melalui pull request, lalu pipeline menjalankan pengujian dan membuat artefak untuk staging." },
    { id:"s3", speaker:"Engineer", start:24, end:41, text:"Persetujuan produksi masih manual. Itu bagian yang paling sering membuat antrean ketika reviewer sedang tidak tersedia." },
    { id:"s4", speaker:"Interviewer", start:41, end:52, text:"Berapa lama biasanya proses dari merge sampai produksi?" },
    { id:"s5", speaker:"Engineer", start:52, end:70, text:"Kalau lancar sekitar tiga puluh menit, tetapi bisa menjadi beberapa jam saat persetujuan tertunda." },
  ]},
  { id: "demo-002", filename: "daily-ops-14-aug.wav", status: "processing", stage: "diarizing", progress: 72, createdAt: "2026-08-14T00:56:00.000Z", createdLabel: "9m lalu", duration: 428 },
];

const fmtTime = (s=0) => `${String(Math.floor(s/60)).padStart(2,"0")}:${String(Math.floor(s%60)).padStart(2,"0")}`;
const ago = (iso:string) => { const m=Math.max(1,Math.round((Date.now()-new Date(iso).getTime())/60000)); return m<60?`${m}m lalu`:`${Math.floor(m/60)}j lalu`; };

export default function Home() {
  const [jobs,setJobs]=useState<Job[]>(seed); const [selected,setSelected]=useState("demo-001");
  const [query,setQuery]=useState(""); const [drag,setDrag]=useState(false); const [uploading,setUploading]=useState(false);
  const input=useRef<HTMLInputElement>(null); const audio=useRef<HTMLAudioElement>(null);
  const current=jobs.find(j=>j.id===selected) || jobs[0];
  const segments=useMemo(()=>current?.segments?.filter(s=>s.text.toLowerCase().includes(query.toLowerCase()))||[],[current,query]);

  useEffect(()=>{ const id=setInterval(()=>setJobs(all=>all.map(j=>j.status==="processing"?{...j,progress:Math.min(96,j.progress+1)}:j)),2400); return()=>clearInterval(id); },[]);
  async function upload(file?:File){ if(!file)return; setUploading(true); const local:Job={id:crypto.randomUUID(),filename:file.name,status:"queued",stage:"waiting",progress:0,createdAt:new Date().toISOString()}; setJobs(x=>[local,...x]); setSelected(local.id);
    try { const form=new FormData(); form.append("audio",file); const res=await fetch(`${API}/jobs`,{method:"POST",body:form}); if(!res.ok)throw new Error(); const remote=await res.json(); setJobs(x=>x.map(j=>j.id===local.id?{...j,id:remote.id}:j)); setSelected(remote.id); } catch { setTimeout(()=>setJobs(x=>x.map(j=>j.id===local.id?{...j,status:"processing",stage:"transcribing",progress:18}:j)),700); } finally {setUploading(false);} }
  function drop(e:DragEvent){e.preventDefault();setDrag(false);upload(e.dataTransfer.files[0]);}
  function fileChange(e:ChangeEvent<HTMLInputElement>){upload(e.target.files?.[0]);e.target.value="";}
  function exportText(){if(!current?.segments)return; const body=current.segments.map(s=>`[${fmtTime(s.start)}–${fmtTime(s.end)}] ${s.speaker}\n${s.text}`).join("\n\n"); const a=document.createElement("a");a.href=URL.createObjectURL(new Blob([body],{type:"text/plain"}));a.download=`${current.filename}.txt`;a.click();URL.revokeObjectURL(a.href);}

  return <main>
    <header><Link className="brand" href="/"><span className="mark">V</span><span>VOXTRACE</span></Link><nav><span className="live"><i/> SYSTEM READY</span><button className="iconbtn" aria-label="Pengaturan">⌘</button></nav></header>
    <section className="hero"><div><p className="eyebrow">AUDIO INTELLIGENCE / WORKSPACE</p><h1>Rekaman menjadi<br/><em>insight terstruktur.</em></h1><p className="lede">Transkripsi presisi, identifikasi pembicara, dan timeline yang dapat ditelusuri—dalam satu ruang kerja.</p></div><div className="stat"><span>PIPELINE</span><strong>Whisper Large</strong><small>+ WhisperX alignment</small><div className="wave">▂▅▃▇▄▆▂▅▇▃▆▂▇▅▃▆▂</div></div></section>
    <section className={`drop ${drag?"drag":""}`} onDragOver={e=>{e.preventDefault();setDrag(true)}} onDragLeave={()=>setDrag(false)} onDrop={drop} onClick={()=>input.current?.click()} role="button" tabIndex={0} onKeyDown={e=>e.key==="Enter"&&input.current?.click()}>
      <input ref={input} type="file" accept="audio/*,.mp3,.wav,.m4a,.flac,.ogg" hidden onChange={fileChange}/><span className="uploadIcon">↑</span><div><strong>{uploading?"Mengunggah…":"Letakkan rekaman baru di sini"}</strong><small>MP3, WAV, M4A, FLAC · Maks. 2 GB</small></div><button>{uploading?"TUNGGU":"PILIH AUDIO"}</button>
    </section>
    <div className="workspace"><aside><div className="sectionTitle"><span>RECENT RECORDINGS</span><b>{jobs.length}</b></div><div className="jobs">{jobs.map(j=><button key={j.id} onClick={()=>setSelected(j.id)} className={`job ${selected===j.id?"active":""}`}><span className={`status ${j.status}`}/><span className="jobcopy"><strong>{j.filename}</strong><small>{j.status==="processing"?`${j.stage} · ${j.progress}%`:j.status==="queued"?"menunggu worker":j.createdLabel??ago(j.createdAt)}</small>{j.status==="processing"&&<i style={{width:`${j.progress}%`}}/>}</span><span>›</span></button>)}</div></aside>
      <article>{current?.status==="completed"?<><div className="articleTop"><div><p className="eyebrow">TRANSCRIPT / {current.id.toUpperCase()}</p><h2>{current.filename}</h2><p>{fmtTime(current.duration)} · Bahasa Indonesia · {current.segments?.length} segmen</p></div><button className="export" onClick={exportText}>↓ EXPORT .TXT</button></div><div className="audio"><button aria-label="Putar audio" onClick={()=>audio.current?.play()}>▶</button><audio ref={audio}><track kind="captions" src="/empty.vtt" srcLang="id" label="Bahasa Indonesia"/></audio><span>00:00</span><div className="track"><i/></div><span>{fmtTime(current.duration)}</span><button aria-label="Kecepatan">1×</button></div><div className="tools"><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Cari dalam transkrip…"/><span>{segments.length} segmen</span></div><div className="transcript">{segments.map(s=><button key={s.id} className="segment" onClick={()=>{if(audio.current)audio.current.currentTime=s.start}}><time>{fmtTime(s.start)}</time><b className={s.speaker==="Engineer"?"violet":""}>{s.speaker}</b><p>{s.text}</p></button>)}</div></>:<div className="processing"><div className="spinner"/><p className="eyebrow">{current?.status?.toUpperCase()}</p><h2>{current?.filename}</h2><p>Pipeline sedang menjalankan tahap <strong>{current?.stage}</strong>.</p><div className="bigprogress"><i style={{width:`${current?.progress||0}%`}}/></div><b>{current?.progress||0}%</b></div>}</article>
    </div><footer><span>VOXTRACE / LOCAL-FIRST</span><span>Audio tetap di infrastruktur Anda</span></footer>
  </main>;
}
