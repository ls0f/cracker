# cracker
proxy over http[s], support http,socks5 proxy.

```
+------------+            +--------------+          
| local app  |  <=======> |local proxy   | <#######
+------------+            +--------------+        #
                                                  #
                                                  #
                                                  # http[s]
                                                  #
                                                  #
+-------------+            +--------------+       #
| target host |  <=======> |http[s] server|  <#####
+-------------+            +--------------+         
```

# Install

Download the latest binaries from this [release page](https://github.com/lovedboy/cracker/releases).

You can also install from source if you have go installed.

```
# on server
go get github.com/lovedboy/cracker/server
# on local
go get github.com/lovedboy/cracker/local
```
# Usage

## Server side (Run on your vps or other application container platform)

```
./server -addr :8080 -secret <password>
```

## Local side (Run on your local pc)

```
./local -raddr http://example.com:8080 -secret <password>
```

## https

It is strongly recommended to open the https option on the server side.

### Notice

The file name of certificate and private key must be `cert.pem` and `key.pem` and with the server bin under the same folder.

If you have a ssl certificate, It would be easy.

copy the certificate and private key into the same folder with server bin

```
./server -addr :8080 -secret <password> -https
```

```
./local -raddr https://example.com:8080 -secret <password>
```

Of Course, you can create a self-signed ssl certificate by openssl.

```
curl https://raw.githubusercontent.com/lovedboy/cracker/master/gen_key_cert.sh | sh
```

```
./server -addr :8080 -secret <password> -https
```
copy the certificate into the same folder with local bin.

```
./local -raddr https://<ip>:8080 -secret <password>
```


# Quick Test

If you don't want to run the server side, I did for you :) you only need to run the local side.

```
./local  -raddr https://lit-citadel-13724.herokuapp.com -secret 123456
```

[Deploy the server size on heroku](https://github.com/lovedboy/cracker-heroku)


# Next

Play with [SwitchyOmega](https://github.com/FelisCatus/SwitchyOmega/releases)

