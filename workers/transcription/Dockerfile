FROM nvidia/cuda:12.8.1-cudnn-runtime-ubuntu22.04
ARG PIP_INDEX_URL=https://pypi.org/simple
ENV DEBIAN_FRONTEND=noninteractive \
    PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True
RUN sed -i 's|http://archive.ubuntu.com|https://archive.ubuntu.com|g; s|http://security.ubuntu.com|https://security.ubuntu.com|g' /etc/apt/sources.list && \
    apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg python3 python3-pip python3-venv && \
    rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY requirements.txt .
RUN --mount=type=cache,target=/root/.cache/pip \
    python3 -m pip install --index-url "$PIP_INDEX_URL" --timeout 300 --retries 10 -r requirements.txt
COPY main.py .
CMD ["python3","main.py"]
