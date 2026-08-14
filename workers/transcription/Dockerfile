FROM nvidia/cuda:12.8.1-cudnn-runtime-ubuntu22.04
ENV DEBIAN_FRONTEND=noninteractive \
    PIP_NO_CACHE_DIR=1 \
    PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg python3 python3-pip python3-venv && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY requirements.txt .
RUN python3 -m pip install -r requirements.txt
COPY main.py .
CMD ["python3","main.py"]
