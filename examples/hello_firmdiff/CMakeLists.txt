cmake_minimum_required(VERSION 3.16)
project(hello_firmdiff C)

add_executable(app main.c)

set_target_properties(app PROPERTIES
  OUTPUT_NAME "app"
  SUFFIX ".elf"
)