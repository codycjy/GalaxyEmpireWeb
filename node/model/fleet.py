from dataclasses_json import dataclass_json
from dataclasses import dataclass
from config import ShipToID


@dataclass_json
@dataclass
class Fleet:
    ds: int = 0       # death star
    de: int = 0       # destroyer
    cargo: int = 0    # cargo ship
    bs: int = 0       # battleship
    satellite: int = 0  # satellite
    lf: int = 0       # light fighter
    hf: int = 0       # heavy fighter
    cr: int = 0       # cruiser
    dr: int = 0       # dreadnought
    bomb: int = 0     # bomber
    guard: int = 0    # guard ship

    # Method 1: Using __dict__ and dict comprehension
    def to_fleet(self):
        return {ShipToID[attr]: value for attr, value in self.__dict__.items()}


if __name__ == '__main__':
    pass
