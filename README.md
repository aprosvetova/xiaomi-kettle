# Xiaomi Kettle BLE protocol
![The kettle](https://avatars.mds.yandex.net/get-mpic/96484/img_id1900574683909775425/9hq)

## Overview
Xiaomi Kettle is based on QN9020 MCU that makes use of GATT over Bluetooth Low Energy to communicate with mobile app.

The protocol allows to control some settings and get status updates.
It's closed so I used my brain and some tools to decompile and analyze the Mi Home app and the native library that handles encryption.

I don't give any guarantees that this will work for you.

## Caveats
I should probably start with bad news.
 - There is NO way to heat up water if Keep Warm mode is off.
 - The ONLY way to enable/reenable Keep Warm mode is to press physical button.
 - Keep Warm mode has a configurable time limit but you can't set it higher than 12 hours until you can't hack encrypted MCU firmware. The mode turns off when time passes.
 - Keep Warm mode turns off if current temperature drops fast. For example, if you have 80°C water and refill the kettle with cold water, the temperature will drop and the mode will turn off.
 - Keep Warm mode turns off if the kettle is low at water.
 - Minimum Keep Warm mode temperature is 40 °C.

## Connection
At first you just need to connect to your kettle via BLE.

You can use any programming language and any BLE library that supports writing/reading characteristics and subscribing them. 

I've used Go and [currantlabs/ble](http://github.com/currantlabs/ble) but I'm not going to publish my code now since it's not perfect.

Successfully connected? Let's authenticate then.

## Authentication
You will need some variables and functions to start:

 1. `reversedMac` is your kettle's address but reversed, 6 bytes. For example, reversedMac for AA:BB:CC:DD:EE:FF is 0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA.
 2. `productID` is always 131 I guess.
 3. `token` is the... token used to authenticate your kettle, 12 bytes! You can generate random token every auth, no matter. If you want to use Mi Home too, then pair Mi Home with your kettle and use the token from it.
 4. `cipher`, `mixA`, `mixB` are functions from Xiaomi native library. They are common for lots of devices and used to cipher auth packets. I have a [Go implementation](cipher.go).
 5. BLE characteristics used below can be found in [Сharacteristics section](#characteristics).

Let's start.

 1. Send `0x90, 0xCA, 0x85, 0xDE` bytes to `authInitCharacteristic`.
 2. Subscribe `authCharacteristic`.
 3. Send `cipher(mixA(reversedMac, productID), token)` to `authCharacteristic`.
 4. Now you'll get a notification on `authCharacteristic`. You must wait for this notification before proceeding to next step. The notification data can be ignored or used to check an integrity, **this is optional**. If you want to perform a check, compare `cipher(mixB(reversedMac, productID), cipher(mixA(reversedMac, productID), res))` where `res` is received payload with your `token`, they must equal.
 5. Send `0x92, 0xAB, 0x54, 0xFA` to `authCharacteristic`.
 6. Read from `verCharacteristics`. You can ignore the response data, you just have to perform a read to complete authentication process.

You are authenticated now, so you can subscribe [status updates](#status-updates) and/or send [commands](#commands).

## Status updates
If you want to receive status updates, you need to subscribe `statusCharacteristic`.
After you subscribe it, you'll start receiving notifications with payloads.
Here is the payload scheme:
|Byte index|Description|Values|
|:---:|--|--|
|0|Action|0 - Idle<br/>1 - Heating<br/>2 - Cooling<br/>3 - Keeping warm|
|1|Mode (corresponding to LEDs)|0 - None<br/>1 - Boil<br/>2 - Keep Warm|
|2-3|Unknown||
|4|Keep Warm set temperature|40-95 in °C|
|5|Current temperature|0-100 in °C|
|6|Keep Warm type|0 - Boil and cool down to set temperature<br/>1 - Just heat up to set temperature|
|7-8|Keep Warm time|Time passed in minutes since keep warm was enabled|

## Commands
|Description|Characteristic|Payload|Readable|Writable|
|--|--|--|:---:|:---:|
|Keep Warm time limit|`timeCharacteristic`|From 0 to 12 hours multiplied by 2.<br/><br/>*7.5 hours is 15, e.g.*|+|+|
|Keep Warm type and temperature|`setupCharacteristic`|First byte: type, 0 or 1<br/><br/>Second byte: temperature, 40-95|-|+|
|Turn off after boil|`boilModeCharacteristic`|0 - No<br/>1 - Yes|+|+|
|Firmware version|`mcuVersionCharacteristic`|string|+|-|

## Characteristics
|Name|UUID|
|--|:---:|
|`authInitCharacteristic`|0010|
|`authCharacteristic`|0001|
|`verCharacteristics`|0004|
|<br/>|<br/>|
|`setupCharacteristic`|aa01|
|`statusCharacteristic`|aa02|
|`timeCharacteristic`|aa04|
|`boilModeCharacteristic`|aa05|
|<br/>|<br/>|
|`mcuVersionCharacteristic`|2a28|

## Usage example
![Home Assistant integration](https://i.imgur.com/fOlTRZ7.png)

I used Xiaomi Kettle protocol to develop kettle<->MQTT bridge in Go that allows me to integrate my kettle with [Home Assistant](http://home-assistant.io).

Leftmost icon shows current temperature.

I can "enable" or "disable" my kettle by toggling the rightmost icon.
Actually, I keep Keep Warm mode always enabled on my kettle so this On just means "set temperature to 85 °C via `setupCharacteristic`", Off means "set temperature to 40 °C" since 40 °C is the minimum.

My Keep Warm type is always  1 (heat up to set temperature without boiling).

When I leave my home or go to sleep, Home Assistant "turns off" kettle automatically so it rests at 40 °C. When I come home or wake up, Home Assistant prepares 85 °C water for me!

I can also tap&hold leftmost icon and it will make my kettle boil the water (no matter if 85 °C mode is "on" or "off").
This is how I make it boil:
 1. Set Keep Warm type to 0 via `setupCharacteristic`.
 2. Wait for "Cooling" action in [status updates](#status-updates).
 3. Set Keep Warm type back to 1.

When it's ready Home Assistant also sends me "The kettle is boiling" Telegram message so I can pick it up.
I don't use my boil feature frequently because my 85 °C setting is usually enough to make tea. I need 100 °C water only when I want some soothing herbs.

Home Assistant has HomeKit component so my kettle is also available with Apple Home app and Siri. I can yell "Hey Siri, boil", "Hey Siri, turn kettle on/off" or "Hey Siri, current kettle temperature" at my HomePod :D.

Of course, as explained in [caveats](#caveats), I always need to turn on my Keep Warm mode by pressing physical button when 12 hour limit passes or when I refill the kettle with cold water, to make all these magical things work.

## Credits
The work is done by Anna Prosvetova.

Thanks to [jadx-gui](https://github.com/skylot/jadx) for APK decompiling, [IDA Pro](https://www.hex-rays.com/products/ida/index.shtml) for native library disassembling and my dear friend [@Scorpi](https://github.com/Scorpi) for moral support and lots of help.

Are you Russian? You can [subscribe my Telegram channel](https://tele.gg/theyforcedme) <3.