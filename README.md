What is this?
-------------

Software for a separate trusted hardware device ("hardware security module") which
essentially acts just like [Keepass](http://keepass.info/) and only serves the
function of storing secrets.

If you use Keepass on your PC and your PC gets compromised by a virus or a hacker,
it's game over.

If you use a separate device for storing secrets, your PC compromise does not
expose your secrets. This software only exposes your secret when you physically
press a button on the device - and only exposes one secret per push acknowledge.


Features
--------

- Supported secrets:
	* Passwords
	* OTP tokens (Google Authenticator)
	* SSH keys (via SSH agent protocol)
- Create, view and list secrets in a folder hierarchy.
- Export database to Keepass format (for viewing in mobile devices when traveling etc.)
- Import data from Keepass format


Recommended hardware
--------------------

Raspberry Pi. I'm using [Zero W](https://www.raspberrypi.org/products/pi-zero-w/)
with [wooden case](https://thepihut.com/products/zebra-zero-for-raspberry-pi-zero-wood)
and a [capacitive pushbutton](http://www.ebay.com/sch/?_nkw=ttp223).



Building
--------

```
$ docker build -f Dockerfile.build -t pi-security-module .
$ docker run --rm -it -p 8080:80 -v "$(pwd):/go/src/github.com/function61/pi-security-module" pi-security-module
$ glide install
$ go build
```

(the Docker build container is optional, but it makes everything easier and doesn't install anything in your host system)

Releasing: take a look at `bin/release.sh`


TODO
----

- Tags to .JS command definitions
- Enter to confirm command dialog
- Data types for command fields (password)
