# Python 3.12.12; install only verified wheels from the resolved dependency lock.
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c
ENV PYTHONDONTWRITEBYTECODE=1
RUN groupadd -r publisher && useradd -r -g publisher -u 1000 publisher
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir --only-binary :all: --require-hashes -r requirements.txt
COPY publisher.py ./
USER publisher
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=3)"
EXPOSE 8000
CMD ["uvicorn", "publisher:app", "--host", "0.0.0.0", "--port", "8000"]
