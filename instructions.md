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

You can parse the game TITLE and TITLE_ID within the file 'PS3_GAME/PARAM.SFO' using 'internal/parsers/param_sfo.go'.

- For selected folders/files, you may need to recursively search inside for the 'PS3_GAME' folder.
- For 7z files that are selected, you will first need to extract the 'PS3_GAME/PARAM.SFO' file to parse game information, and efficiently extract all files/folders only within the folder containing 'PS3_GAME' into the output directory.
- Some games may just be a '.pkg' file, you can ignore those for now.
- Ensure that non-file characters in the game TITLE are not inserted into the directory. Remove ':', etc. from the TITLE.
- By default you can compress the 'game/' folder with 7z, but if flag '--decompressed' is passed, you can leave the 'game/' folder uncompressed.


Example Usage:

```bash
ps3-game-packager pack roms/ps3/* -o roms/ps3/organized/ 
```

Input:

```
'3D Dot Game Heroes (USA).zip'
'Army of Two (USA)'/
'Battlefield - Bad Company (USA) (Gold Edition)'/
'Brothers in Arms - Hell'\''s Highway (USA)'/
```

Output:

```
'3D DOT GAME HEROES [BLUS30490]'/
'Army of Two (TM) [BLUS30057]'/
'BATTLEFIELD Bad Company™ [BLUS30121]'/
'Brothers in Arms Hell'\''s Highway™ [BLUS30165]'/
```
