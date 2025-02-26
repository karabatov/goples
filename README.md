# Goples

Goples is a small Windows tray app that calls a Discord channel webhook when it detects that the specified process is running.

It checks if the process is running once a minute, but only sends the message once while the process is continuously running. If you exit the process and start it again, it will send a new message.

## Installation

Download `goples.zip` from the Releases tab. 

Unpack the archive somewhere. Inside is a folder called `goples` with `goples.exe` and `icon.ico` inside.

Make a shortcut for `goples.exe` and add the following to its command line:

```
goples.exe -url "<Discord webhook URL>" -process "<game.exe>" -message "I'm playing game.exe!"
```

Run the shortcut. The app should appear in your tray. Right-click to exit.