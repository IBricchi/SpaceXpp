/*
  Program written by Benjamin Ramhorst (br319@ic.ac.uk) during May-June 2021
  as part of SpaceX++ team for the Summer term design project at Imperial College London
  SMPS code originally written by Yue Zhu (yue.zhu18@imperial.ac.uk) in July 2020 and further
  modfied by Benjamin Ramhorst to fit the requirements of the Mars Rover
  Teamplete code for optical flow and movement functions taken from MS Teams for the Summer Electronics Design Project 2 2021
  and further modified by Benjamin Ramhorst to fit the project requirements
  UART code written by Nicholas Pfaff and modified by Benjamin Ramhorst
*/

#include "smps.h"

// Current sensing chip
INA219_WE ina219;

// Duty cycle required to achieve the desired voltage
float duty_cycle;

// Measurements from the circuit - output voltage and inductor current
float vout, iL;

// Constants
const float Ts = 0.0008; // [s]
const float kpv = 0.05024, kiv = 15.78, kdv = 0;
const float uv_max = 4, uv_min = 0;
const float kpi = 0.02512, kii = 39.4, kdi = 0;
const float ui_max = 1, ui_min = 0;
const float current_limit = 3.0;

// Internal values needed for the PID controller
// u->output; e->error; 0->this time; 1->last time; 2->last last time
float u0v, u1v, delta_uv, e0v, e1v, e2v;

// Internal values for the current controller
// u->output; e->error; 0->this time; 1->last time; 2->last last time
float u0i, u1i, delta_ui, e0i, e1i, e2i;

// This subroutine processes all of the analogue samples, creating the required values for the main loop
void sampling() {
  vout = (float)(analogRead(A0)) * (4.096 / 1023.0);
  iL = (float)(ina219.getCurrent_mA()) / 1000.0;
}

// Saturation function - used for limiting current/voltage to avoid damaging the circuit
float saturation(float sat_input, float uplim, float lowlim) {
  if (sat_input > uplim) sat_input = uplim;
  else if (sat_input < lowlim ) sat_input = lowlim;
  return sat_input;
}

// PWM waveform for SMPS duty cycle
void pwm_modulate(float pwm_input) {
  analogWrite(PIN_PWM_OUT, (int)(255 - pwm_input * 255));
}

// Voltage PID controller
float pidv(float pid_input) {
  float e_integration;
  e0v = pid_input;
  e_integration = e0v;
  if (u1v >= uv_max) {
    e_integration = 0;
  } else if (u1v <= uv_min) {
    e_integration = 0;
  }
  delta_uv = kpv * (e0v - e1v) + kiv * Ts * e_integration + kdv / Ts * (e0v - 2 * e1v + e2v);
  u0v = u1v + delta_uv;  //this time's control output
  saturation(u0v, uv_max, uv_min);
  u1v = u0v;
  e2v = e1v;
  e1v = e0v;
  return u0v;
}

// Current PID Controller
float pidi(float pid_input) {
  float e_integration;
  e0i = pid_input;
  e_integration = e0i;
  if (u1i >= ui_max) {
    e_integration = 0;
  } else if (u1i <= ui_min) {
    e_integration = 0;
  }
  delta_ui = kpi * (e0i - e1i) + kii * Ts * e_integration + kdi / Ts * (e0i - 2 * e1i + e2i);
  u0i = u1i + delta_ui;  //this time's control output
  saturation(u0i, ui_max, ui_min);
  u1i = u0i;
  e2i = e1i;
  e1i = e0i;
  return u0i;
}

// Basic setup for the SMPS pins and current sensing chip
void SMPSSetup() {
  // Need this for millis to work
  interrupts();

  // External analogue reference for the ADC
  analogReference(EXTERNAL);

  // TimerB0 initialization for PWM output
  pinMode(PIN_PWM_OUT, OUTPUT);
  TCB0.CTRLA = TCB_CLKSEL_CLKDIV1_gc | TCB_ENABLE_bm;
  analogWrite(PIN_PWM_OUT, 120);

  // Current sensor
  Wire.begin();
  ina219.init();
  Wire.setClock(700000);
}

// A function for controlling the SMPS duty cycle to achieve the desired reference voltage
void SMPSControl(float vref) {
  /*
    The closed loop path has a voltage controller cascaded with a current controller. The voltage controller
    creates a current demand based upon the voltage error. This demand is saturated to give current limiting.
    The current loop then gives a duty cycle demand based upon the error between demanded current and measured
    current.
  */
  sampling();
  float ev = vref - vout;
  float cv = pidv(ev);
  cv = saturation(cv, current_limit, 0);
  float ei = cv - iL;
  duty_cycle = pidi(ei);
  duty_cycle = saturation(duty_cycle, 0.99, 0.01);
  pwm_modulate(duty_cycle);
}
