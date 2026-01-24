FROM python:3.11-slim

RUN groupadd -r subscriber && useradd -r -g subscriber -u 1000 subscriber

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY subscriber.py .

RUN chown -R subscriber:subscriber /app

USER subscriber

EXPOSE 9090

CMD ["python", "subscriber.py"]
