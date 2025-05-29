
docker stop prometheus-vastai-exporter-1
docker rm prometheus-vastai-exporter-1
docker build -t frznfngrs/vastai-exporter .
docker push frznfngrs/vastai-exporter
cd /home/vast/prometheus/
docker-compose up -d
cd /home/vast/prometheus-vastai

