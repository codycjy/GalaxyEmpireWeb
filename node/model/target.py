from dataclasses import dataclass
from dataclasses_json import dataclass_json


@dataclass_json
@dataclass
class Target:
    galaxy: int
    system: int
    planet: int
    is_moon = False

    def to_dict(self):
        return {
            "galaxy": self.galaxy,
            "system": self.system,
            "planet": self.planet,
        }
