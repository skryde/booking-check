import asyncio
import logging
import os
import time
import tomllib
from enum import Enum

import telegram
from selenium import webdriver
from selenium.webdriver.common.by import By

logger = logging.getLogger(__name__)
screenshot_file_name = 'screenshot.png'


class Result(Enum):
    FOUND = 1
    NOT_FOUND = 2
    ERROR = 3


def do_web_scraping() -> Result:
    options = webdriver.FirefoxOptions()
    options.add_argument("-headless")  # Commenting this line could be helpful for testing the script.

    driver = webdriver.Firefox(options=options)
    result = Result.NOT_FOUND

    try:
        driver.get(
            "https://www.exteriores.gob.es/Consulados/montevideo/es/ServiciosConsulares/Paginas/index.aspx?scco=Uruguay&scd=200&scca=Pasaportes+y+otros+documentos&scs=Pasaportes+-+Requisitos+y+procedimiento+para+obtenerlo")
        driver.implicitly_wait(90)
        link = driver.find_element(By.LINK_TEXT, "Cita Pasaportes")
        link.click()

        handles = driver.window_handles
        driver.switch_to.window(handles[1])

        default_container = driver.find_element(By.ID, "idBktDefaultServicesContainer")
        while not default_container.is_displayed():
            time.sleep(0.5)

        div = driver.find_element(By.ID, "idDivBktServicesContainer")
        divs = div.find_elements(By.TAG_NAME, "div")

        text = "No hay horas disponibles.\nInténtelo de nuevo dentro de unos días."

        for d in divs:
            if d.is_displayed() and d.text == text:
                result = Result.FOUND

        driver.save_screenshot(screenshot_file_name)

    except Exception as e:
        result = Result.ERROR
        logger.error("error in scraping process", exc_info=e)

    finally:
        driver.quit()

    return result


async def send_message(bot_token: str, msg: str, recipients: list[str]):
    bot = telegram.Bot(token=bot_token)
    for recipient in recipients:
        await bot.send_message(recipient, msg)

        try:
            os.stat(screenshot_file_name)
            await bot.send_photo(recipient, screenshot_file_name)

        except FileNotFoundError:
            logger.warning("there is no screenshot to send")


if __name__ == '__main__':
    log_format = '%(asctime)s %(levelname)s [%(name)s]: %(message)s'
    logging.basicConfig(filename='booking-check.log', format=log_format, level=logging.INFO)

    try:
        with open("config.toml", "rb") as config_file:
            config = tomllib.load(config_file)

    except FileNotFoundError:
        logger.error("'config.toml' configuration file not found")
        exit(1)

    tg_bot_api_token = config.get("telegram_bot_api_token")
    if len(tg_bot_api_token) < 1:
        logger.error("the telegram bot api token seems not valid in the configuration file 'config.toml'")
        exit(1)

    message_recipients = config.get("message_recipients")
    if len(message_recipients) < 1:
        logger.error("no recipients present in the configuration file 'config.toml'")
        exit(1)

    # Default result will be ERROR.
    r = Result.ERROR

    try:
        # This try/except will catch the error that could occur when creating the Firefox driver.
        r = do_web_scraping()

    except Exception as e:
        logger.error("error on 'do_web_scraping()'", exc_info=e)
        exit(1)

    if r == Result.FOUND:
        logger.info("there are no available hours")

    elif r == Result.ERROR:
        message = "Error validating hour availability"
        asyncio.run(send_message(tg_bot_api_token, message, message_recipients))

    else:
        message = "There are hours available"
        asyncio.run(send_message(tg_bot_api_token, message, message_recipients))
