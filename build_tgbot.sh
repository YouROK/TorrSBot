#!/bin/bash

git clone --recursive https://github.com/tdlib/telegram-bot-api.git tgbotapi
mkdir tgbotapi/build
cd tgbotapi/build
cmake -DCMAKE_BUILD_TYPE=Release ..
make -j8
cp telegram-bot-api ../../dist