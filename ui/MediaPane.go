package ui

import (
	"image"
	"math"
	"sync"
	"time"

	"github.com/ninjasphere/driver-go-gestic/gestic"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/rpc"
)

type MediaPane struct {
	log *logger.Logger
	rpc *rpc.Client

	lastAirWheelTime time.Time
	lastAirWheel     *uint8

	volume      float64
	volumeImage *Image
	muteImage   *Image

	gestureSync *sync.Mutex

	controlDevices []string
	volumeDevices  []string
}

func NewMediaPane(volumeImage string, muteImage string, rpcClient *rpc.Client, thingType string) *MediaPane {
	log := logger.GetLogger("MediaPane")

	controlDevices, err := getChannelIds(thingType, "media-control", rpcClient)
	if err != nil {
		log.Fatalf("Failed to get %s devices: %s", err, err)
	}
	log.Infof("Pane got %d media-control devices", len(controlDevices))

	volumeDevices, err := getChannelIds(thingType, "media-control", rpcClient)
	if err != nil {
		log.Fatalf("Failed to get %s devices: %s", err, err)
	}
	log.Infof("Pane got %d volume devices", len(volumeDevices))

	return &MediaPane{
		log:              log,
		volumeDevices:    volumeDevices,
		controlDevices:   controlDevices,
		rpc:              rpcClient,
		gestureSync:      &sync.Mutex{},
		volumeImage:      loadImage(volumeImage),
		muteImage:        loadImage(muteImage),
		lastAirWheelTime: time.Now(),
	}
}

func (p *MediaPane) Gesture(gesture *gestic.GestureData) {
	p.gestureSync.Lock()

	if gesture.AirWheel.AirWheelVal > 0 {

		if time.Since(p.lastAirWheelTime) > time.Millisecond*300 {
			p.lastAirWheel = nil
		}

		p.lastAirWheelTime = time.Now()

		//p.log.Debugf("Airwheel: %d", gesture.AirWheel.AirWheelVal)

		if p.lastAirWheel != nil {
			offset := int(gesture.AirWheel.AirWheelVal) - int(*p.lastAirWheel)

			if offset > 30 {
				offset -= 255
			}

			if offset < -30 {
				offset += 255
			}

			//p.log.Debugf("Airwheel New: %d Offset: %d Last: %d", gesture.AirWheel.AirWheelVal, offset, *p.lastAirWheel)

			//p.log.Debugf("Current volume %f", p.volume)

			//p.log.Debugf("Volume offset %f", float64(offset)/255.0)

			var volume float64 = p.volume + float64(offset)/255.0

			//p.log.Debugf("Adjusted volume %f:", volume)

			volume = math.Max(volume, 0)
			volume = math.Min(volume, 1)

			if p.volume != volume {
				p.log.Debugf("New volume %f:", volume)
				p.volume = volume
			}
		}

		val := gesture.AirWheel.AirWheelVal
		p.lastAirWheel = &val
		//spew.Dump("last2", p.lastAirWheel)
	}

	p.gestureSync.Unlock()
}

func (p *MediaPane) Render() (*image.RGBA, error) {
	if p.volume > 0 {
		return p.volumeImage.GetPositionFrame(1 - p.volume), nil
	}

	return p.muteImage.GetNextFrame(), nil
}

func (p *MediaPane) IsDirty() bool {
	return true
}