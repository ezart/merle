## Quick Start, Part II

In Part I, we ran Merle locally on a Raspberry Pi and used a web browser to view the Thing on http://localhost.  In Part II, we're going to run Merle on another system on the Internet.  This second system runs a copy of the Thing.  We'll call this copy Thing' (Thing prime).  Thing and Thing' are syncronized.  Any state change in Thing is reflected in Thing', and visa-versa.  Thing' view through a web browser shows the same web-app view we see when viewing Thing.  In fact, the same front-end code (HTML, CSS, Javascript) and back-end code (Go) is running in both places.  But notice Thing' does not have access to hardware.  Thing' is a proxy for Thing's hardware.  

<img src="https://docs.google.com/drawings/d/e/2PACX-1vTnTWN_AcEEWgGjuN7uF13Z3lCikFMPzGR7eSovbkdgLS0__YcJ5Azh6BWQZLchuh12HZahR8VyH05F/pub?w=2023&amp;h=921">

Thing and Thing' are connected to each other over a secure SSH tunnel.  To create the tunnel, we'll need to generate and install an SSH key.  On the raspberry Pi system, generate a new key:

```
admin@pi $ ssh-keygen
Generating public/private rsa key pair.
Enter file in which to save the key (/home/admin/.ssh/id_rsa): /home/admin/.ssh/mykey
Enter passphrase (empty for no passphrase): 
Enter same passphrase again: 
Your identification has been saved in /home/admin/.ssh/mykey.
Your public key has been saved in /home/admin/.ssh/mykey.pub.
The key fingerprint is:
SHA256:gujwAWq4fCm6CPHWAhv9pql7o6CD6ZgoSndsLVXlHGA admin@pi
The key's randomart image is:
+---[RSA 2048]----+
|         E.o     |
|        . + .    |
|.        . o     |
|oo . .  .        |
|B.+ . ..S        |
|o@ +o o.         |
|*=*++= .         |
|#o*=o .          |
|^Oo.             |
+----[SHA256]-----+
```

This will generate a public/private rsa key pair.  Give a file name for the key and press return twice for an empty passphrase.  In this example, /home/admin/.ssh/mykey is used as the file name.  mykey.pub holds the public key.  mykey holds the private key.  The user is admin.

Next, we install the public key on the second system, the system on the internet that's going to run Thing'.  In general, we can install the key using: 

```sh
admin@pi $ ssh-copy-id -i [identity file] username@remote_host
```

Where identity file in our example is /home/admin/.ssh/mykey.pub.  For this Quick Start, we can use the same Raspberry Pi system to run Thing and Thing' using localhost as the remote_host address.   Substitute your internet host address for remote_host rather than using localhost.  We're using localhost here to quickly get Thing' running.  Let's install the key:

```
admin@pi $ ssh-copy-id -i /home/admin/.ssh/mykey.pub admin@localhost
```



