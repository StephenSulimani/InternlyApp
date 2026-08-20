package main

import (
	"github.com/stephensulimani/internlyapp/cmd/engine/modules"
	"github.com/stephensulimani/internlyapp/cmd/engine/modules/linkedin"
	"github.com/stephensulimani/internlyapp/internal/ats"
	"github.com/stephensulimani/internlyapp/internal/service"
)

var (
	_ service.JobSource = (*modules.Simplify)(nil)
	_ service.JobSource = (*linkedin.LinkedIn)(nil)
	_ service.JobSource = (*ats.BoardSource)(nil)
	_ ats.Provider      = (*ats.Greenhouse)(nil)
)
