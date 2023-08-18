
docker stop prometheus-vastai-exporter-1
docker rm prometheus-vastai-exporter-1
docker build -t jjziets/vastai-exporter .
docker push jjziets/vastai-exporter
cd /home/vast/prometheus/
docker-compose up -d
cd /home/vast/prometheus-vastai

