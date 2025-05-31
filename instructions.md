Hello, can you help me create a Golang CLI script that will be used to package my legally obtained PS3 roms that are in the format of decrypted ISO. There will be 2 commands, 'pack', and 'unpack', we will just focus on 'pack' for now.

'pack' will be used to convert a decrypted ISO game folder or archive file to the following format:

```
{Game Name} [{Game ID}]
|-- game.7z (when decompressed, will have 'game/PS3_GAME')
|-- _updates (atm empty folder with all update files)
`-- _dlc (atm empty folder with all dlc files + .rap)
```

As an example:
```
3D DOT GAME HEROES [BLUS30490]
|-- game.7z
|-- _updates
`-- _dlc
```

You can find the game TITLE and TITLE_ID within the file 'PS3_GAME/PARAM.SFO' using the provided script 'PS3Dec R5/Windows/PS3Dec.exe'
