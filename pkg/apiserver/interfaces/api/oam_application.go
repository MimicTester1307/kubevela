/*
Copyright 2021 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"k8s.io/klog/v2"

	"github.com/oam-dev/kubevela/pkg/apiserver/domain/service"
	apis "github.com/oam-dev/kubevela/pkg/apiserver/interfaces/api/dto/v1"
	"github.com/oam-dev/kubevela/pkg/apiserver/utils/bcode"
)

type oamApplication struct {
	OamApplicationService service.OAMApplicationService `inject:""`
	RbacService           service.RBACService           `inject:""`
}

// NewOAMApplication new oam application
func NewOAMApplication() Interface {
	return &oamApplication{}
}

func (c *oamApplication) GetWebServiceRoute() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path("/v1").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML).
		Doc("api for oam application manage")

	tags := []string{"oam_application"}

	ws.Route(ws.GET("/namespaces/{namespace}/applications/{appname}").To(c.getApplication).
		Doc("get the specified oam application in the specified namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Filter(c.RbacService.CheckPerm("application", "detail")).
		Param(ws.PathParameter("namespace", "identifier of the namespace").DataType("string")).
		Param(ws.PathParameter("appname", "identifier of the oam application").DataType("string")).
		Returns(200, "OK", apis.ApplicationResponse{}).
		Writes(apis.ApplicationResponse{}))

	ws.Route(ws.POST("/namespaces/{namespace}/applications/{appname}").To(c.createOrUpdateApplication).
		Doc("create or update oam application in the specified namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Filter(c.RbacService.CheckPerm("application", "deploy")).
		Param(ws.PathParameter("namespace", "identifier of the namespace").DataType("string")).
		Param(ws.PathParameter("appname", "identifier of the oam application").DataType("string")).
		Param(ws.QueryParameter("dryRun", "When present, indicates that modifications should not be persisted. "+
			"An invalid or unrecognized dryRun directive will result in an error response and no further processing of the request. "+
			"Valid values are: - All: all dry run stages will be processed").DataType("string").Required(false)).
		Reads(apis.ApplicationRequest{}))

	ws.Route(ws.DELETE("/namespaces/{namespace}/applications/{appname}").To(c.deleteApplication).
		Operation("deleteOAMApplication").
		Doc("create or update oam application in the specified namespace").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Filter(c.RbacService.CheckPerm("application", "delete")).
		Param(ws.PathParameter("namespace", "identifier of the namespace").DataType("string")).
		Param(ws.PathParameter("appname", "identifier of the oam application").DataType("string")))

	ws.Filter(authCheckFilter)
	return ws
}

func (c *oamApplication) getApplication(req *restful.Request, res *restful.Response) {
	namespace := req.PathParameter("namespace")
	appName := req.PathParameter("appname")
	appRes, err := c.OamApplicationService.GetOAMApplication(req.Request.Context(), appName, namespace)
	if err != nil {
		klog.Errorf("get application failure %s", err.Error())
		bcode.ReturnError(req, res, err)
		return
	}

	// Write back response data
	if err := res.WriteEntity(appRes); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
}

func (c *oamApplication) createOrUpdateApplication(req *restful.Request, res *restful.Response) {
	namespace := req.PathParameter("namespace")
	appName := req.PathParameter("appname")

	var createReq apis.ApplicationRequest
	if err := req.ReadEntity(&createReq); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}

	dryRun := req.QueryParameter("dryRun")
	if len(dryRun) != 0 {
		if dryRun != "All" {
			bcode.ReturnError(req, res, bcode.ErrApplicationDryRunFailed.SetMessage("Invalid dryRun parameter. Must be 'All'"))
			return
		}
		if err := c.OamApplicationService.DryRunOAMApplication(req.Request.Context(), createReq, appName, namespace); err != nil {
			klog.Errorf("dryrun application failure %s", err.Error())
			bcode.ReturnError(req, res, err)
			return
		}
	} else {
		if err := c.OamApplicationService.CreateOrUpdateOAMApplication(req.Request.Context(), createReq, appName, namespace); err != nil {
			klog.Errorf("create application failure %s", err.Error())
			bcode.ReturnError(req, res, err)
			return
		}
	}

	if err := res.WriteEntity(apis.EmptyResponse{}); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
}

func (c *oamApplication) deleteApplication(req *restful.Request, res *restful.Response) {
	namespace := req.PathParameter("namespace")
	appName := req.PathParameter("appname")

	err := c.OamApplicationService.DeleteOAMApplication(req.Request.Context(), appName, namespace)
	if err != nil {
		klog.Errorf("delete application failure %s", err.Error())
		bcode.ReturnError(req, res, err)
		return
	}

	if err := res.WriteEntity(apis.EmptyResponse{}); err != nil {
		bcode.ReturnError(req, res, err)
		return
	}
}
