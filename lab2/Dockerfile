FROM python:3.11-slim

WORKDIR /app

COPY server_not_rate_limit.py .
COPY server_race.py .
COPY server_rate_limit.py .
COPY server_single_thread.py .
COPY client.py .
COPY test_conc.py .
COPY test_rate.py .

RUN mkdir -p /app/content

EXPOSE 8080

CMD ["python", "server_not_rate_limit.py", "/app/content"]