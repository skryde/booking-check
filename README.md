# Spanish Consulate in Uruguay | Booking Availability Checker  

## About  

Am I over-engineering this? Yes, probably.  

Why? Because I can 😌 (and it's a great opportunity to learn new technologies).  

Which technologies? NATS, for example.  

### What is this?  

This is a service that checks the availability of appointment slots in the Spanish Consulate in Uruguay's booking system every five minutes.  

If any slots are available, the service notifies all subscribed users via a Telegram Bot message.  

Something like this:  

![image](https://github.com/user-attachments/assets/d81cb9cf-8999-4369-8c65-b8975c13c7da)  

## Getting Started  

### Self-Hosted  

1. Clone or download this repository.  

2. Run: `cp .env.example .env`  

3. Edit the `.env` file and set your Telegram Bot token and your Telegram ID. Your Telegram ID is used to grant you, as the bot owner/admin, access to extra commands that are not available to everyone.  

   If you don't know your Telegram ID, you can run the bot without setting the `TELEGRAM_BOT_OWNER_ID` variable and retrieve your ID by executing the bot command `/me`.  

4. Start the containers: `docker compose up -d`  

## Admin Commands  

- `/status`

  Returns the debug status (`true` or `false`) along with the list of subscribed user IDs.

  Example:

  ```
  System status:

  Debug status: false
  Subscriptions: [12345678 12345679]
  ```

- `/enabledebug`

  Enables debug messages. This means that the bot will send every scraping result to the admin/owner, regardless of whether the admin is subscribed or not.  

- `/disabledebug`

  Disables debug messages.
