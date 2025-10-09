FROM python:3.11-slim

WORKDIR /app

COPY server.py .
COPY client.py .

RUN mkdir -p /app/content

EXPOSE 8080

CMD ["python", "server.py", "/app/content"]