#!/usr/bin/env bash

openssl req -subj '/CN=*/' -x509 -sha256 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 1024 -nodes
