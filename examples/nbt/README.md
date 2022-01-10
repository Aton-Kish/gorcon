# NBT

| Environment Varriables | Default         |
| ---------------------- | --------------- |
| RCON_ADDRESS           | minecraft:25575 |
| RCON_PASSWORD          | minecraft       |

If you want to configure:

```bash
export RCON_ADDRESS=...
export RCON_PASSWORD=...
```

### Execute

```bash
go run main.go
```

### Output

```text
Player: YourName (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
NBT   : {"AbsorptionAmount":0,"Air":300,"Attributes":[{"Base":0.10000000149011612,"Name":"minecraft:generic.movement_speed"}], ... ,"seenCredits":0}
Player: ... (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
NBT   : { ... }
Player: ... (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
...
```

For details about Player NBT format, see <https://minecraft.fandom.com/wiki/Player.dat_format>.
