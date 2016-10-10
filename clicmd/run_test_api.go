package clicmd

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

// RunTestAPI runs the a sample backend API for which the Agent can be
// developed against.
func RunTestAPI(c *cli.Context) {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		IndentJSON: true, // Output human readable JSON
	}))
	m.Get("/api", func(r render.Render) {
		staticResponse := map[string]interface{}{
			"cluster": map[string]interface{}{
				"name":  "cluster-name",
				"scope": "cluster-scope",
			},
			"wale_mode": "aws",
			"wale_env": []string{
				// fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", os.Getenv("AWS_ACCESS_KEY_ID")),
				// fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", os.Getenv("AWS_SECRET_ACCESS_KEY")),
				// fmt.Sprintf("WAL_S3_BUCKET=%s", os.Getenv("WAL_S3_BUCKET")),
				// fmt.Sprintf("WALE_S3_ENDPOINT=%s", os.Getenv("WALE_S3_ENDPOINT")),
				fmt.Sprintf("AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY_ID"),
				fmt.Sprintf("AWS_SECRET_ACCESS_KEY=AWS_SECRET_ACCESS_KEY"),
				fmt.Sprintf("WAL_S3_BUCKET=WAL_S3_BUCKET"),
				fmt.Sprintf("WALE_S3_ENDPOINT=WALE_S3_ENDPOINT"),
			},
			"postgresql": map[string]interface{}{
				"admin": map[string]interface{}{
					"username": "admin-username",
					"password": "admin-password",
				},
				"superuser": map[string]interface{}{
					"username": "superuser-username",
					"password": "superuser-password",
				},
				"appuser": map[string]interface{}{
					"username": "appuser-username",
					"password": "appuser-password",
				},
			},
			"etcd": map[string]interface{}{
				"uri": os.Getenv("ETCD_URI"),
			},
		}
		r.JSON(200, staticResponse)
	})
	m.Run()

}