# Python 3.12.12; install only verified wheels from the resolved dependency lock.
FROM python@sha256:f3fa41d74a768c2fce8016b98c191ae8c1bacd8f1152870a3f9f87d350920b7c
ENV PYTHONDONTWRITEBYTECODE=1
RUN groupadd -r batchuser && useradd -r -g batchuser -u 1000 batchuser
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir --only-binary :all: --require-hashes -r requirements.txt
COPY batch_processor.py ./
USER batchuser
CMD ["python", "batch_processor.py"]
