import asyncio
import base64
import json
import logging
import os
import time
import uuid
from enum import Enum

import nats
from selenium import webdriver
from selenium.common import NoAlertPresentException, TimeoutException
from selenium.webdriver.common.by import By
from selenium.webdriver.firefox.service import Service as FirefoxService
from selenium.webdriver.support.wait import WebDriverWait
from webdriver_manager.firefox import GeckoDriverManager

logger = logging.getLogger(__name__)
screenshot_file_name = 'screenshot.png'


class Result:
    def __init__(self):
        self.status = ResultStatus.NOT_FOUND
        self.message = ''


class ResultStatus(Enum):
    FOUND = 1
    NOT_FOUND = 2
    ERROR = 3


def do_web_scraping() -> Result:
    try:
        os.stat(screenshot_file_name)
        os.remove(screenshot_file_name)

    except FileNotFoundError:
        logger.debug("there is no screenshot to remove")

    options = webdriver.FirefoxOptions()
    options.add_argument("-headless")  # Commenting this line could be helpful for testing the script.

    exe_path = GeckoDriverManager().install()
    driver = webdriver.Firefox(service=FirefoxService(executable_path=exe_path), options=options)
    result = Result()

    try:
        driver.get(
            "https://www.exteriores.gob.es/Consulados/montevideo/es/ServiciosConsulares/Paginas/index.aspx?scco=Uruguay&scd=200&scca=Pasaportes+y+otros+documentos&scs=Pasaportes+-+Requisitos+y+procedimiento+para+obtenerlo")
        driver.implicitly_wait(90)
        link = driver.find_element(By.LINK_TEXT, "Cita Pasaportes")
        link.click()

        handles = driver.window_handles
        driver.switch_to.window(handles[1])

        errors = [NoAlertPresentException]
        wait = WebDriverWait(driver, timeout=10, poll_frequency=.2, ignored_exceptions=errors)
        wait.until(lambda foo: driver.switch_to.alert.accept() or True)

        btn = driver.find_element(By.ID, "idCaptchaButton")
        btn.click()

        default_container = driver.find_element(By.ID, "idBktDefaultServicesContainer")
        while not default_container.is_displayed():
            time.sleep(0.5)

        div = driver.find_element(By.ID, "idDivBktServicesContainer")
        divs = div.find_elements(By.TAG_NAME, "div")

        text = "No hay horas disponibles.\nInténtelo de nuevo dentro de unos días."

        for d in divs:
            if d.is_displayed() and d.text == text:
                result.status = ResultStatus.FOUND

    except TimeoutException as scraping_exception:
        result.status = ResultStatus.ERROR
        result.message = "timeout accessing 'Cita Pasaportes' page"
        logger.error("error in scraping process", exc_info=scraping_exception)

    except Exception as scraping_exception:
        result.status = ResultStatus.ERROR
        result.message = "unhandled error"
        logger.error("error in scraping process", exc_info=scraping_exception)

    finally:
        try:
            driver.save_screenshot(screenshot_file_name)

        except Exception as screenshot_exception:
            logger.warning("error taking browser window screenshot", exc_info=screenshot_exception)

        driver.quit()

    return result


async def notify(nats_host: str, msg: str, debug: bool) -> None:
    b64 = bytes()
    try:
        os.stat(screenshot_file_name)
        image = open(screenshot_file_name, 'rb')
        b64 = base64.b64encode(image.read())

    except FileNotFoundError:
        logger.warning("there is no screenshot to send")

    nc = await nats.connect(nats_host)

    a = {
        "debug": debug,
        "message": msg,
        "image": b64.decode("utf-8"),
    }
    j = json.dumps(a)

    await nc.publish("scrapper.result", j.encode("utf-8"))
    await nc.flush()
    await nc.close()


if __name__ == '__main__':
    log_format = '%(asctime)s %(levelname)s [' + str(uuid.uuid4()) + '] [%(name)s]: %(message)s'
    logging.basicConfig(filename='booking-check.log', format=log_format, level=logging.INFO)

    nats_server_host = os.getenv('NATS_HOST', 'nats://127.0.0.1:4222')

    try:
        # This try/except will catch the error that could occur when creating the Firefox driver.
        r = do_web_scraping()

    except Exception as e:
        logger.error("error on 'do_web_scraping()'", exc_info=e)
        exit(1)

    if r.status == ResultStatus.FOUND:
        # No hours available text found in the booking webpage.
        logger.info("there are no available hours")
        asyncio.run(
            notify(nats_server_host, "There are no available hours", True))

    elif r.status == ResultStatus.ERROR:
        message = "Error validating hour availability: " + r.message
        asyncio.run(notify(nats_server_host, message, True))

    else:
        asyncio.run(notify(nats_server_host, "There are hours available", False))
