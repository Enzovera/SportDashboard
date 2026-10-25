FROM python:3.13-slim AS base
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_ROOT_USER_ACTION=ignore PIP_NO_CACHE_DIR=1
WORKDIR /app
COPY pyproject.toml ./
COPY ingest ./ingest
RUN pip install . \
 && useradd --uid 10001 --create-home ingest \
 && mkdir -p /data/garmin /data/fitatu && chown ingest /data/garmin /data/fitatu && chmod 700 /data/garmin /data/fitatu
USER ingest
ENTRYPOINT ["python", "-m", "ingest"]

FROM base AS dev
USER root
RUN pip install ".[dev]"
USER ingest
