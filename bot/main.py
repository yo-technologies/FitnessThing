import signal
from telegram.ext import Application, ContextTypes, CommandHandler, MessageHandler, filters
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, Update, WebAppInfo
from dotenv import load_dotenv
import logging
import os
import asyncio

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
    """Handle the /start command - welcome message with web app access."""
    webapp = WebAppInfo(url=url)
    keyboard = InlineKeyboardMarkup(
        [
            [
                InlineKeyboardButton("🏋️ Открыть приложение", web_app=webapp),
            ]
        ]
    )
    welcome_message = (
        "🏋️‍♀️ Добро пожаловать в FitnessThing!\n\n"
        "Я ваш персональный фитнес-помощник. Здесь вы можете:\n"
        "• Получить персональные тренировки\n"
        "• Отслеживать прогресс\n"
        "• Управлять целями\n\n"
        "Нажмите кнопку ниже, чтобы открыть приложение, или используйте /help для получения списка команд."
    )
    await update.message.reply_text(welcome_message, reply_markup=keyboard)

async def help_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /help command - show available commands."""
    help_text = (
        "🤖 <b>Доступные команды:</b>\n\n"
        "/start - Главное меню и доступ к приложению\n"
        "/help - Показать это сообщение\n"
        "/about - Информация о боте\n\n"
        "💡 <b>Основные функции:</b>\n"
        "• Отправьте любое сообщение, и я повторю его\n"
        "• Используйте кнопку в /start для доступа к полному приложению\n"
        "• В приложении доступны персональные тренировки и отслеживание прогресса"
    )
    await update.message.reply_text(help_text, parse_mode='HTML')

async def about_command(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Handle the /about command - information about the bot."""
    about_text = (
        "🏋️‍♀️ <b>FitnessThing Bot</b>\n\n"
        "Я персональный фитнес-помощник, созданный для помощи в достижении ваших фитнес-целей.\n\n"
        "🎯 <b>Что я умею:</b>\n"
        "• Предоставляю доступ к веб-приложению с персональными тренировками\n"
        "• Генерирую индивидуальные программы тренировок с помощью ИИ\n"
        "• Помогаю отслеживать прогресс\n"
        "• Адаптируюсь под ваш уровень подготовки\n\n"
        "📱 Для полного функционала используйте веб-приложение через команду /start"
    )
    await update.message.reply_text(about_text, parse_mode='HTML')

async def echo(update: Update, context: ContextTypes.DEFAULT_TYPE):
    """Echo user message back - handle non-command messages."""
    if update.message and update.message.text:
        echo_text = f"🔄 Вы написали: {update.message.text}\n\nИспользуйте /help для списка команд или /start для доступа к приложению."
        await update.message.reply_text(echo_text)
    else:
        logger.warning("Received an update without a message")

async def main():
    app = Application.builder().token(os.getenv("TG_BOT_TOKEN")).build()

    # Register command handlers
    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("help", help_command))
    app.add_handler(CommandHandler("about", about_command))
    
    # Register message handler for non-command messages (echo functionality)
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, echo))

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
    asyncio.run(main())
