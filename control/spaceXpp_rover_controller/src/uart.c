#include "uart.h"

void drive_uart_init()
{
    const uart_config_t drive_uart_config = {
        .baud_rate = 115200, // Arduino could go higher than this (see datasheet)
        .data_bits = UART_DATA_8_BITS,
        .parity = UART_PARITY_DISABLE,
        .stop_bits = UART_STOP_BITS_1,
        .flow_ctrl = UART_HW_FLOWCTRL_DISABLE, // not supported by standard Arduino
    };
    ESP_ERROR_CHECK(uart_param_config(DRIVE_UART_NUM, &drive_uart_config));

    ESP_ERROR_CHECK(uart_set_pin(DRIVE_UART_NUM, DRIVE_TXD_PIN, DRIVE_RXD_PIN, UART_PIN_NO_CHANGE, UART_PIN_NO_CHANGE));

    ESP_ERROR_CHECK(uart_driver_install(DRIVE_UART_NUM, DRIVE_BUFFER_SIZE, DRIVE_BUFFER_SIZE, 0, NULL, 0));
}

void vision_uart_init()
{
    // TODO: implement based on DE10-Lite UART interface
}

void drive_uart_task(void *arg)
{
    uint8_t* data = (uint8_t*) malloc(DRIVE_BUFFER_SIZE+1);
    while (1) {
        int rxBytes = uart_read_bytes(DRIVE_UART_NUM, data, DRIVE_BUFFER_SIZE, 1000 / portTICK_RATE_MS);
        if (rxBytes > 0) {
            data[rxBytes] = 0; // End of received string
            printf("UART data from drive: %s", (char*)data);

            // Send string data
            char* tx_string = "A message from ESP32\n";
            uint8_t stringKey = (uint8_t)MSG;
            uart_write_bytes(UART_NUM_1, (const char*)&stringKey, 1); // send key
            uart_write_bytes(UART_NUM_1, (const char*)tx_string, strlen(tx_string)); // send data
        }

         vTaskDelay(1000 / portTICK_PERIOD_MS);
    }
    free(data);
}

void vision_uart_task(void *arg)
{
    while (1) {
        // TODO: implement
    }
}
