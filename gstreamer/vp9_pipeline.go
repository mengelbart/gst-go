package gstreamer

type VP9RTPPipeline struct {
	*Pipeline
}

func NewVP9ToRTPPipeline(src *Element, eosCB EOSHandler, errCB ErrorHandler) (*VP9RTPPipeline, error) {
	builders := Elements{
		src,
		NewElement("vp9enc",
			Set("name", "encoder"),
			Set("error-resilient", "default"),
			Set("keyframe-max-dist", 10),
			Set("cpu-used", 4),
			Set("deadline", 1),
		),
		NewElement("rtpvp9pay",
			Set("name", "rtpvp9pay"),
			Set("mtu", 1500),
			Set("seqnum-offset", 0),
		),
		NewElement("appsink",
			Set("name", "appsink"),
		),
	}

	p, err := NewPipeline(builders.Build(), eosCB, errCB)
	if err != nil {
		return nil, err
	}
	return &VP9RTPPipeline{p}, nil
}

type RTPToVP9Pipeline struct {
}
