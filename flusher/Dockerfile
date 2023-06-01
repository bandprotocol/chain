FROM python:3.8-slim

COPY . .
RUN apt-get update && apt-get install -y libpq-dev gcc
RUN pip install -r requirements.txt

# TODO: Using environment variable for db and kafka bootstrap services
CMD python main.py sync --db postgres:postgrespassword@postgres:5432/postgres -s kafka:9092
