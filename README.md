# vastai_exporter

![Docker Image CI](https://github.com/500farm/prometheus-vastai/actions/workflows/docker-image.yml/badge.svg)

For [Vast.ai](https://vast.ai) hosts.

Prometheus exporter reporting data from your Vast.ai account:

- Stats of your machines: reliability, DLPerf score, inet speed, number of client jobs running, number of gpus used.
- Stats of your own instances: on-demand and default.
- Paid and pending balance of your account.
- Your on-demand and bid prices. 
- Stats of hosts' offerings of GPU models that you have.

In per-account Prometheus metrics at  (url: `/metrics`), 

_NOTE: This is a work in progress. Output format is subject to change._

### Usage

```
docker run -d --restart always -p 8622:8622 jjziets/vastai-exporter \
    --api-key=VASTKEY
```
Replace _VASTKEY_ with your Vast.ai API key. To test, open http://localhost:8622. If does not work, check container output with `docker logs`.


