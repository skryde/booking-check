# Booking Check

## Installation
```shell
git clone https://github.com/skryde/booking-check.git
cd booking-check

./install.sh # This will create a Python 'venv' under que .venv directory

echo 'telegram_bot_api_token = "bot<token>"' > config.toml # Change 'bot<token>' with your Telegram Bot API Token.
echo 'message_recipients = ["123456"]' >> config.toml # Change '123456' with your Telegram User ID.

./run.sh
```

## Crontab configuration
```
# Execute every 5 minutes.
*/5 * * * * /path/to/booking-check/run.sh
```