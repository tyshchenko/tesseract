/*  Copyright 2019 The tesseract Authors

    This file is part of tesseract.

    tesseract is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    tesseract is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package tesseract

// TODO: the central regions of the Milky Way are heavily obscured by dust.
//
// https://en.wikipedia.org/wiki/Baade%27s_Window
// https://en.wikipedia.org/wiki/Dark_nebula
//
// https://en.wikipedia.org/wiki/Molecular_cloud
// https://en.wikipedia.org/wiki/Open_cluster
// https://en.wikipedia.org/wiki/H_II_region
//
import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// Sector is a galactic 3D cube volume used by the procedural generation of
// stars and other galactic-level features.
type Sector struct {
	// Galactic X,Y,Z in gridUnit
	Corner *V3

	// Value between 0.0 and 1.0 denoting how much the sector has been
	// explored by players == procedurally generated
	Mapped float64

	// Star system positions are in gridUnit
	StarSystems []*StarSystem
}

func (s *Sector) Clone() interface{} {
	sss := []*StarSystem{}
	for _, ss := range s.StarSystems {
		sss = append(sss, ss.Clone().(*StarSystem))
	}
	return &Sector{
		Corner:      s.Corner.Clone().(*V3),
		Mapped:      s.Mapped,
		StarSystems: sss,
	}
}

// Traverse enacts partial procedural generation of a sector.
// Traverse is generally called by the TODO system when a player is traversing
// the sector (having spent X time and/or moved Y distance in the sector).
func (s *Sector) Traverse() {
	if s.Mapped == 1.0 {
		return
	}

	if s.Mapped == 0.0 {
		s.getAdjacents() // ensures adjacent sectors are initialized
	}

	prob := (s.stellarDensity() / StarMassAvg) * sectorTraversalFactor
	if Rand.Float64() < prob {
		ss := NewStarSystem(StarMassHist.Sample())
		s.addStarSystem(ss)
	}

	s.Mapped += sectorTraversalFactor
	if s.Mapped > 1.0 {
		s.Mapped = 1.0
	}
}

func (s *Sector) addStarSystemFixed(newStarSystem *StarSystem, pos *V3) {
	rf := &RefFrame{
		Parent:      rootRF,
		Pos:         pos,
		Orbit:       nil,
		Orientation: nil, // TODO
	}
	S.EntFrames[newStarSystem.Entity] = rf

	for _, p := range newStarSystem.Planets {
		prf := &RefFrame{
			Parent:      rf,
			Pos:         nil,
			Orbit:       p.Body.Orbit,
			Orientation: nil, // TODO
		}
		S.EntFrames[p.Entity] = prf
	}

	s.StarSystems = append(s.StarSystems, newStarSystem)
	S.AddStarSystem(newStarSystem, pos)
}

func (s *Sector) addStarSystem(newStarSystem *StarSystem) {
	p := &V3{}
	randDist := func() float64 {
		return (sectorSize/gridUnit)*Rand.Float64() - 1.0
	}
	randPos := func() {
		p.X = s.Corner.X + randDist()
		p.Y = s.Corner.Y + randDist()
		p.Z = s.Corner.Z + randDist()
	}

	sectors := append(s.getAdjacents(), s)
	x := new(V3)

NewPos:
	randPos()
	for _, sec := range sectors {
		//fmt.Printf("%s\n", sec.Debug())
		//sec.Debug()
		for _, ss := range sec.StarSystems {
			pos := S.Pos[ss.Entity]
			diff := x.Sub(pos, p).Magnitude()
			//fmt.Printf("diff: %.2f\n", diff)
			if diff < minStellarProximity {
				//fmt.Println("NewPos")
				goto NewPos
			}
		}
	}
	s.addStarSystemFixed(newStarSystem, p)
}

// getAdjacents returns the 26 cube sectors adjacent to s
func (s *Sector) getAdjacents() []*Sector {
	type P struct{ X, Y, Z int }
	permutations := []P{
		// all permutations with Z up
		P{0, 0, 1},
		P{1, 0, 1},
		P{0, 1, 1},
		P{1, 1, 1},

		P{-1, 0, 1},
		P{0, -1, 1},
		P{-1, -1, 1},

		P{1, -1, 1},
		P{-1, 1, 1},

		// all permutations with Z down
		P{0, 0, -1},
		P{1, 0, -1},
		P{0, 1, -1},
		P{1, 1, -1},

		P{-1, 0, -1},
		P{0, -1, -1},
		P{-1, -1, -1},

		P{1, -1, -1},
		P{-1, 1, -1},

		// all permutations with Z zero (except self)
		// P{0, 0, 0},
		P{1, 0, 0},
		P{0, 1, 0},
		P{1, 1, 0},

		P{-1, 0, 0},
		P{0, -1, 0},
		P{-1, -1, 0},

		P{1, -1, 0},
		P{-1, 1, 0},
	}

	adjacents := make([]*Sector, 26)
	a, b := new(V3), new(V3)
	for i := 0; i < 26; i++ {
		b.X = auToGrid(float64(permutations[i].X))
		b.Y = auToGrid(float64(permutations[i].Y))
		b.Z = auToGrid(float64(permutations[i].Z))
		a.Add(s.Corner, b)
		adjacents[i] = GetSector(a)
	}

	return adjacents
}

func GetSector(pos *V3) *Sector {
	key := sectorKey(pos, true)
	s, ok := S.Sectors[key]
	if ok {
		return s
	} else {
		newP := &V3{}
		newP.X, newP.Y, newP.Z = pos.X, pos.Y, pos.Z
		newS := &Sector{newP, 0.0, make([]*StarSystem, 0)}
		S.Sectors[key] = newS
		return newS
	}
}

func (s *Sector) Key() string {
	return sectorKey(s.Corner, true)
}

func sectorKey(pos *V3, floor bool) string {
	//log.Debug("FUNKY", "pos", pos)
	format := func(f float64) string {
		u := f / (sectorSize / gridUnit)
		if floor {
			u = math.Floor(u)
		}
		//fmt.Printf("sectorKey f, floor, calc: %.9f, %.9f, %.9f\n", f, floor, calc)
		return strconv.FormatFloat(u, 'f', 2, 64)
	}
	return "[" +
		format(pos.X) + "," +
		format(pos.Y) + "," +
		format(pos.Z) + "]"
}

// See: https://arxiv.org/pdf/1811.07911.pdf
//
// TODO: use a more fitting function for p (M☉pc-3) (Fig. 4. in section 5.2)
// TODO: find a source for abs(Z) > 200 pc
//
// TODO: model milky way spiral arm position and densities
// TODO: model galactic core position and density
//       (using coordinate system defined in section 2)
func (s *Sector) stellarDensity() float64 {
	// how far above/below we are to the galactic plane (mid plane)
	depth := math.Abs(s.Corner.Z) / (sectorSize / gridUnit)
	// no star density beyond the disc
	lim := milkyWayDiscHeight / 2.0
	// see Fig. 3. TODO: more fitting function
	starFraction := 0.6

	var density float64
	switch {
	case depth < 200.0:
		density = 0.2 - depth/(200.0/0.15)
	case depth < lim:
		density = 0.05 - depth/(1000.0/0.05)
	default:
		density = 0.0
	}

	testAddition := 16.0
	return starFraction*density + testAddition
}

func NewGalaxy() {
	// Simulate starship in hyperdrive: stars are detected by proximity
	// but their star systems remain unmapped.

	// Discovery through travel: every X AU of hyperdrive position is checked.
	// Galactic "quadrants" - cubes with side X ly, entering a new one reveals
	// new star systems, continously from being in the cube, until they are all
	// mapped.  Their systems remain unmapped, we only know the base attributes
	// of each star.

	// When entering a cube, X% of its stars are instantly revealed.
	// Then an additional X% every Y minutes.
}

//
// Debug
//
func (s *Sector) Debug() {
	fmt.Printf("%s: mapped: %.2f star count: %d\n", s.Key(), s.Mapped, len(s.StarSystems))
}

func DebugSectors(onlyMapped bool) {
	keys := []string{}
	for key, s := range S.Sectors {
		if onlyMapped && s.Mapped == 0.0 {
			continue
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, k := range keys {
		s := S.Sectors[k]
		fmt.Printf("%s SECTOR: %.2f mapped\n", k, s.Mapped)
		for _, ss := range s.StarSystems {
			fmt.Printf("%s Star System: %.2f mapped\n",
				sectorKey(S.Pos[ss.Star.Entity], false),
				ss.Mapped,
			)
			fmt.Printf("  %s: %s Class (%.2f M☉) %.2f mapped\n",
				ss.Star.Body.Name,
				string(ss.Star.SpectralType),
				ss.Star.Body.Mass,
				ss.Mapped,
			)
			for i, p := range ss.Planets {
				fmt.Printf("    Planet %d: %s (%.2f M⊕)\n", i+1, p.Body.Name, p.Body.Mass)
			}

		}
		fmt.Printf("\n")
	}

}

func auToGrid(au float64) float64 {
	return (au * sectorSize) / gridUnit
}
