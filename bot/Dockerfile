FROM python:3.12-slim

WORKDIR /app

# Устанавливаем зависимости
RUN pip install --no-cache-dir python-telegram-bot python-dotenv

# Копируем исходный код
COPY main.py .

CMD ["python", "main.py"]
