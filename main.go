/*
 * Bamboo - A Vietnamese Input method editor
 * Copyright (C) 2018 Luong Thanh Lam <ltlam93@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	ibus "github.com/BambooEngine/goibus"
)

const (
	ComponentName = "org.freedesktop.IBus.bamboo"
	EngineName    = "Bamboo"
)

var embedded = flag.Bool("ibus", false, "Run the embedded ibus component")
var version = flag.Bool("version", false, "Show version")
var gui = flag.Bool("gui", false, "Show GUI")
var isWayland = false
var isGnome = false
var useManualIntrospect = false

func hasGnome(env string) bool {
	return strings.Contains(strings.ToLower(os.Getenv(env)), "gnome")
}

type Introspector int

const (
	IntrospectorX11 Introspector = iota
	IntrospectorGnome
	IntrospectorWayland
	IntrospectorFifo
)

var introspector = IntrospectorX11

func guessIntrospector() Introspector {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return IntrospectorWayland
	}
	if hasGnome("XDG_CURRENT_DESKTOP") || hasGnome("DESKTOP_SESSION") || hasGnome("GDMSESSION") {
		return IntrospectorGnome
	}
	return IntrospectorX11
}

func main() {

	/// Reading the introspector type
	var forceIntrospector = os.Getenv("BAMBOO_INTROSPECTOR")
	fmt.Printf("Introspector: (%s)\n", forceIntrospector)
	switch forceIntrospector {
	case "wayland":
		introspector = IntrospectorWayland
	case "gnome":
		introspector = IntrospectorGnome
	case "x11":
		introspector = IntrospectorX11
	case "fifo":
		introspector = IntrospectorFifo
	case "auto":
		fallthrough
	case "":
		introspector = guessIntrospector()
	default:
		fmt.Printf("Warning, unknown introspector type (%s)\n", forceIntrospector)
		introspector = guessIntrospector()
	}

	flag.Parse()
	if *embedded {
		os.Chdir(DataDir)
	}

	// Custom initialization for
	// some of the introspector that requires it.
	switch introspector {
	case IntrospectorWayland:
		go wlGetFocusWindowClass()
	case IntrospectorFifo:
		go fifoWatchInitialize()
	}

	if *version {
		fmt.Println(Version)
	} else if *embedded {
		engine := GetIBusEngineCreator()
		bus := ibus.NewBus()
		bus.RequestName(ComponentName, 0)

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, engine)

		select {}
	} else {
		log.SetFlags(log.LstdFlags | log.Lmicroseconds)
		bus := ibus.NewBus()
		log.Println("Got Bus, Running Standalone")
		component := &ibus.Component{
			Name:          "IBusComponent",
			ComponentName: ComponentName + "Standalone",
		}
		engine := &ibus.EngineDesc{
			Name:       "IBusEngineDesc",
			EngineName: EngineName + "Standalone",
		}
		component.AddEngine(engine)
		bus.RegisterComponent(component)

		conn := bus.GetDbusConn()
		ibus.NewFactory(conn, GetIBusEngineCreator())

		bus.CallMethod("SetGlobalEngine", 0, EngineName+"Standalone")

		select {}
	}
}
