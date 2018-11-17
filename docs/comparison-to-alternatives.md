
| Application        | Free            | No cloud | Hacker safe | U2F support | SSH agent support |
|--------------------|-----------------|----------|-------------|-------------|-------------------|
| pi-security-module | ☑              | ☑      | ☑         | ☑         | ☑               |
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

If you notice any errors in this comparison, please report an issue! Being accurate and
honest is important to us.
