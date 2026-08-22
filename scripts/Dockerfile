FROM python:3.12-alpine
RUN apk add --no-cache postgresql-client
WORKDIR /app
COPY seed_found.py /app/seed_found.py
ENV SEED_IN_COMPOSE=1
ENV IMG_DIR=/tmp/seed-images
CMD ["python3", "/app/seed_found.py"]
