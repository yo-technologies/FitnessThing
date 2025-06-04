import signal
from telegram.ext import Application, ContextTypes, CommandHandler
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update, WebAppInfo
from dotenv import load_dotenv
import logging
import os

logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)

logging.getLogger("telegram").setLevel(logging.WARNING)
logging.getLogger("httpx").setLevel(logging.WARNING)
logging.getLogger("httpcore").setLevel(logging.WARNING)

load_dotenv()

url = os.getenv("APP_URL")

class GracefulKiller:
    kill_now = False

    def __init__(self):
        signal.signal(signal.SIGINT, self.exit_gracefully)
        signal.signal(signal.SIGTERM, self.exit_gracefully)

    def exit_gracefully(self, signum, frame):
        logger.info("Exiting gracefully")
        self.kill_now = True

async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    webapp = WebAppInfo(url=url)
    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton("Открыть приложение", web_app=webapp),
            ]
        ]
    )
    await update.message.reply_text("Hello, world!", reply_markup=keyboard)

async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if update.message:
        await update.message.reply_text(update.message.text)
    else:
        logger.warning("Received an update without a message")

async def main():
    app = Application.builder().token(os.getenv("TG_BOT_TOKEN")).build()

    app.add_handler(CommandHandler("start", start))

    await app.initialize()
    await app.start()
    await app.updater.start_polling()

    logger.info("Bot started")
    killer = GracefulKiller()
    while not killer.kill_now:
        await asyncio.sleep(1)

    await app.updater.stop()
    await app.stop()
    await app.shutdown()


if __name__ == "__main__":
    import asyncio
    asyncio.run(main())
