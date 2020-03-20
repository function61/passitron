
| Application        | Free            | No cloud | Hacker safe | U2F support | SSH agent support |
|--------------------|-----------------|----------|-------------|-------------|-------------------|
| Passitron          | ☑              | ☑      | ☑         | ☑         | ☑               |
| KeePass            | ☑              | ☑      | ☐          | ?           | ☑ (plugin)      |
| LastPass           | ☑              | ☐       | ☐          | ?           | ☐                |
| 1Password          | ☐               | ☐       | ☐          | ?           | ☐               |
| Dashlane           | 50 account limit | ☐       | ☐          | ?           | ☐                |

"Hacker safe": if hacker gains access to your computer, can the application still guarantee
that the hacker doesn't get access to all your credentials? Basically the question boils
down to if the application implements BOTH:

1. Does the password database run in a separate machine (with cloud-based apps this is always yes)
2. Are the secrets accessible without a press of a physical button that the hacker doesn't 
   have access to? (i.e. a signature from a U2F key)


On the radar
------------

.. means mention-worthy projects for which we haven't had the time to write a feature
breakdown yet.

[Krypton](https://krypt.co/) - while Krypton is not a password manager (at all), on the
surface (I haven't tested it yet) it seems to do a great job of handling the use case of
U2F-enabled logins. It seems that you need a browser plugin to use it, which is kind of a
bummer - I wonder if it is technically possible to emulate U2F token as a kernel driver to
support every already-U2F-enabled browser without a plugin.

[pass](https://www.passwordstore.org/)

[Bitwarden](https://github.com/bitwarden/core)

[Buttercup](https://buttercup.pw/)


Errors
------

If you notice any errors in this comparison, please report an issue! Being accurate and
honest is important to us.
