/*
    Written by Nicholas Pfaff (nicholas.pfaff19@imperial.ac.uk), 2021 - SpaceX++ EEE/EIE 2nd year group project, Imperial College London
*/

#include <string.h>
#include "esp_log.h"
#include "freertos/FreeRTOS.h"
#include "freertos/queue.h"

#ifndef DRIVE_H
#define DRIVE_H

// tile with in cm
#define MAP_TILE_WIDTH (30)

#define MAX_DRIVE_INSTRUCTION_SEQUENCE_LENGTH (50)
#define DRIVE_INSTRUCTION_DELIMITER ":"

// UART data encoding
typedef struct DriveEncoding{
    const char* forward;
    const char* backward;
    const char* turnRight;
    const char* turnLeft;
    const char* stop;
    const char* stopFromForward;
    const char* stopFromTurn;
} DriveEncoding;

void drive_init();

void flush_drive_instruction_queue();

void drive_backwards_to_last_valid_tile(int distanceDrivenOfLastInstruction);

#endif
