# Goples

Goples is a small Windows tray app that calls a Discord channel webhook when it detects that the specified program is running.

It checks once a minute if the program runs and only sends the message once as long as the program keeps running. If you exit the program and start it again, Goples will send a new message.

Discord does not need to be open for this to work.

## Installation

Download `goples.zip` from the Releases tab. 

Unpack the archive somewhere. Inside is a folder called `goples` with `goples.exe` and `icon.ico` inside.

Make a shortcut for `goples.exe` and add the following to its command line:

```
goples.exe -url "<Discord webhook URL>" -process "<game.exe>" -message "I'm playing game.exe!"
```

Run the shortcut. The app should appear in your tray. Right-click to exit.
