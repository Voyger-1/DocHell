package textfinder

type TextRun struct {
	Text      string
	Bold      bool
	Italic    bool
	Underline bool
	FontSize  int
	Font      string
	Color     string
}

type TextBox struct {
	Runs            []TextRun
	X, Y, Cx, Cy    int64
	ShapeID         string
	ShapeName       string
	BackgroundColor string
	BorderColor     string
	ShapeType       string
}

type slideXML struct {
	SpTree           spTree `xml:"cSld>spTree"`
	AlternateContent []struct {
		Choice struct {
			Sp sp `xml:"sp"`
		} `xml:"Choice>sp"`
		Fallback struct {
			Sp sp `xml:"sp"`
		} `xml:"Fallback>sp"`
	} `xml:"AlternateContent"`
}

type spTree struct {
	Sp           []sp           `xml:"sp"`
	GrpSp        []grpSp        `xml:"grpSp"`
	GraphicFrame []graphicFrame `xml:"graphicFrame"`
}

type sp struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvSpPr>cNvPr"`
	SpPr struct {
		Xfrm struct {
			Off struct {
				X string `xml:"x,attr"`
				Y string `xml:"y,attr"`
			} `xml:"off"`
			Ext struct {
				Cx string `xml:"cx,attr"`
				Cy string `xml:"cy,attr"`
			} `xml:"ext"`
		} `xml:"xfrm"`
		PrstGeom struct {
			Prst string `xml:"prst,attr"`
		} `xml:"prstGeom"`
		SolidFill *struct {
			SrgbClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"srgbClr"`
			SchemeClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"schemeClr"`
		} `xml:"solidFill"`
		Ln *struct {
			SolidFill *struct {
				SrgbClr *struct {
					Val string `xml:"val,attr"`
				} `xml:"srgbClr"`
				SchemeClr *struct {
					Val string `xml:"val,attr"`
				} `xml:"schemeClr"`
			} `xml:"solidFill"`
		} `xml:"ln"`
	} `xml:"spPr"`
	TxBody *txBody `xml:"txBody"`
}

type grpSp struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvGrpSpPr>cNvPr"`
	GrpSpPr struct {
		Xfrm xfrm `xml:"xfrm"`
	} `xml:"grpSpPr"`
	Sp           []sp           `xml:"sp"`
	GrpSp        []grpSp        `xml:"grpSp"`
	GraphicFrame []graphicFrame `xml:"graphicFrame"`
}

type graphicFrame struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvGraphicFramePr>cNvPr"`
	Xfrm    xfrm `xml:"xfrm"`
	Graphic *struct {
		GraphicData *struct {
			Tbl *table `xml:"tbl"`
		} `xml:"graphicData"`
	} `xml:"graphic"`
}

type table struct {
	Tr []struct {
		Tc []struct {
			TxBody *txBody `xml:"txBody"`
		} `xml:"tc"`
	} `xml:"tr"`
}

type txBody struct {
	LstStyle *struct {
		Lvl1pPr *struct {
			DefRPr *rPr `xml:"defRPr"`
		} `xml:"lvl1pPr"`
		Lvl2pPr *struct {
			DefRPr *rPr `xml:"lvl2pPr"`
		} `xml:"lvl2pPr"`
		Lvl3pPr *struct {
			DefRPr *rPr `xml:"lvl3pPr"`
		} `xml:"lvl3pPr"`
		Lvl4pPr *struct {
			DefRPr *rPr `xml:"lvl4pPr"`
		} `xml:"lvl4pPr"`
		Lvl5pPr *struct {
			DefRPr *rPr `xml:"lvl5pPr"`
		} `xml:"lvl5pPr"`
		Lvl6pPr *struct {
			DefRPr *rPr `xml:"lvl6pPr"`
		} `xml:"lvl6pPr"`
		Lvl7pPr *struct {
			DefRPr *rPr `xml:"lvl7pPr"`
		} `xml:"lvl7pPr"`
		Lvl8pPr *struct {
			DefRPr *rPr `xml:"lvl8pPr"`
		} `xml:"lvl8pPr"`
		Lvl9pPr *struct {
			DefRPr *rPr `xml:"lvl9pPr"`
		} `xml:"lvl9pPr"`
	} `xml:"lstStyle"`
	Para []struct {
		PPr *struct {
			DefRPr *rPr   `xml:"defRPr"`
			Lvl    string `xml:"lvl,attr"`
		} `xml:"pPr"`
		R []struct {
			RPr *rPr   `xml:"rPr"`
			T   string `xml:"t"`
		} `xml:"r"`
		AR []struct {
			RPr *rPr   `xml:"rPr"`
			T   string `xml:"t"`
		} `xml:"a:r"`
	} `xml:"p"`
}

type rPr struct {
	Sz    string `xml:"sz,attr"`
	B     string `xml:"b,attr"`
	I     string `xml:"i,attr"`
	U     string `xml:"u,attr"`
	Latin *struct {
		Typeface string `xml:"typeface,attr"`
	} `xml:"latin"`
	SolidFill *struct {
		SrgbClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"srgbClr"`
		SchemeClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"schemeClr"`
		PrstClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"prstClr"`
	} `xml:"solidFill"`
}

type xfrm struct {
	Off struct {
		X string `xml:"x,attr"`
		Y string `xml:"y,attr"`
	} `xml:"off"`
	Ext struct {
		Cx string `xml:"cx,attr"`
		Cy string `xml:"cy,attr"`
	} `xml:"ext"`
}
