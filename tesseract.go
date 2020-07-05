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

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/log"
)

//
// References:
//
// [1] Millington, Ian. Game physics engine development (Second Edition). CRC Press, 2010.
// [2] https://github.com/idmillington/cyclone-physics/blob/master/include/cyclone/core.h
//

func init() {
	log.Root().SetHandler(log.MultiHandler(
		log.StreamHandler(os.Stderr, log.TerminalFormat(true)),
		log.LvlFilterHandler(
			log.LvlDebug,
			log.Must.FileHandler("tesseract_errors.json", log.JSONFormat()))))

	StarMassHist, StarMassAvg = getStarStats()
	PlanetMassHist, PlanetRadiusHist = getExoplanetHistograms()

	fmt.Printf("StarMassAvg: %.4f \n", StarMassAvg)

	//InitWorld()
}
