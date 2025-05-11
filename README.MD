Инструкция по запуску и проверке

1. Запуск брокеров, ui, sr
docker compose -f docker-compose.yml up -d
2. Команда создания топика
docker compose exec kafka-0 kafka-topics.sh --create --topic topic --bootstrap-server kafka-0:9092 --partitions 3 --replication-factor 2 
3. Команда создания образа для продюсера и консьюмеров
docker compose -f docker-compose-create-image.yaml up -d
4. Команда поднятия консьюмеров и продюсера
docker compose -f docker-compose-up-kafka-reader-writer.yaml up -d
