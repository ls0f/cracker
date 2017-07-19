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

If you have a ssl certificate, It would be easy.

```
./server -addr :443 -secret <password> -https -cert /etc/cert.pem -key /etc/key.pem
```

```
./local -raddr https://example.com -secret <password>
```

Of Course, you can create a self-signed ssl certificate by openssl.

```
sh -c "$(curl https://raw.githubusercontent.com/lovedboy/cracker/master/gen_key_cert.sh)"
```

```
./server -addr :443 -secret <password> -https -cert /etc/self-signed-cert.pem -key /etc/self-ca-key.pem
```

```
./local -raddr https://example.com -secret <password> -cert /etc/self-signed-cert.pem
```


# Quick Test

If you don't want to run the server side, I did for you :) you only need to run the local side.

```
./local  -raddr https://lit-citadel-13724.herokuapp.com -secret 123456
```

[Deploy the server side on heroku](https://github.com/lovedboy/cracker-heroku)


# Next

Play with [SwitchyOmega](https://github.com/FelisCatus/SwitchyOmega/releases)
