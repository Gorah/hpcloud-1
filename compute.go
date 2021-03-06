package hpcloud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

/* Server flavours Smallest to Largest */
type Flavor int

const (
	XSmall = Flavor(100) + iota
	Small
	Medium
	Large
	XLarge
	DblXLarge
)

/* Available images */
type ServerImage int

const (
	UbuntuLucid10_04Kernel    = ServerImage(1235)
	UbuntuLucid10_04          = 1236
	UbuntuMaverick10_10Kernel = 1237
	UbuntuMaverick10_10       = 1238
	UbuntuNatty11_04Kernel    = 1239
	UbuntuNatty11_04          = 1240
	UbuntuOneiric11_10        = 5579
	UbuntuPrecise12_04        = 8419
	CentOS5_8Server64         = 54021
	CentOS6_2Server64Kernel   = 1356
	CentOS6_2Server64Ramdisk  = 1357
	CentOS6_2Server64         = 1358
	DebianSqueeze6_0_3Kernel  = 1359
	DebianSqueeze6_0_3Ramdisk = 1360
	DebianSqueeze6_0_3Server  = 1361
	Fedora16Server64          = 16291
	BitNamiDrupal7_14_0       = 22729
	BitNamiWebPack1_2_0       = 22731
	BitNamiDevPack1_0_0       = 4654
	ActiveStateStackatov1_2_6 = 14345
	ActiveStateStackatov2_2_2 = 59297
	ActiveStateStackatov2_2_3 = 60815
	EnterpriseDBPPAS9_1_2     = 9953
	EnterpriseDBPSQL9_1_3     = 9995
)

type Link struct {
	HREF string `json:"href"`
	Rel  string `json:"rel"`
}

/*
  Several embedded types are simply an ID string with a slice of Link
*/
type IDLink struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Links []Link `json:"links"`
}

type Image struct {
	I struct {
		Name     string            `json:"name"`
		ID       string            `json:"id"`
		Links    []Link            `json:"links"`
		Progress int               `json:"progress"`
		Metadata map[string]string `json:"metadata"`
		Status   string            `json:"status"`
		Updated  string            `json:"updated"`
	} `json:"image"`
}

type Images struct {
	I []IDLink `json:"images"`
}

/*
  This type describes the JSON data which should be sent to the create
  server resource.
*/
type Server struct {
	ConfigDrive    bool              `json:"config_drive"`
	FlavorRef      Flavor            `json:"flavorRef"`
	ImageRef       ServerImage       `json:"imageRef"`
	MaxCount       int               `json:"max_count"`
	MinCount       int               `json:"min_count"`
	Name           string            `json:"name"`
	Key            string            `json:"key_name"`
	Personality    string            `json:"personality"`
	UserData       string            `json:"user_data"`
	SecurityGroups []IDLink          `json:"security_groups"`
	Metadata       map[string]string `json:"metadata"`
}

/*
  This type describes the JSON response from a successful CreateServer
  call.
*/
type ServerResponse struct {
	S struct {
		Status         string            `json:"status"`
		Updated        string            `json:"update"`
		HostID         string            `json:"hostId"`
		UserID         string            `json:"user_id"`
		Name           string            `json:"name"`
		Links          []Link            `json:"links"`
		Addresses      interface{}       `json:"addresses"`
		TenantID       string            `json:"tenant_id"`
		Image          IDLink            `json:"image"`
		Created        string            `json:"created"`
		UUID           string            `json:"uuid"`
		AccessIPv4     string            `json:"accessIPv4"`
		AccessIPv6     string            `json:"accessIPv6"`
		KeyName        string            `json:"key_name"`
		AdminPass      string            `json:"adminPass"`
		Flavor         IDLink            `json:"flavor"`
		ConfigDrive    string            `json:"config_drive"`
		ID             int64             `json:"id"`
		SecurityGroups []IDLink          `json:"security_groups"`
		Metadata       map[string]string `json:"metadata"`
	} `json:"server"`
}

/*
  CreateServer creates a new server in the HPCloud using the
  settings found in the Server instance passed to this function.

  This function implements the interface as described in:-
  * https://docs.hpcloud.com/api/compute/
  * section 4.4.5.2 Create Server
*/
func (a Access) CreateServer(s Server) (*ServerResponse, error) {
	b, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}

	body, err := a.baseComputeRequest("servers", "POST",
		strings.NewReader(string(b)),
	)
	if err != nil {
		return nil, err
	}
	sr := &ServerResponse{}
	err = json.Unmarshal(body, sr)
	if err != nil {
		return nil, err
	}
	return sr, nil
}

/*
  DeleteServer deletes the server with the `server_id`.

  This function implements the interface described in:-
  * https://docs.hpcloud.com/api/compute/
  * Section 4.4.6.3 Delete Server
*/
func (a Access) DeleteServer(server_id string) error {
	_, err := a.baseComputeRequest(
		fmt.Sprintf("servers/%s", server_id),
		"DELETE", nil,
	)
	if err != nil {
		return err
	}
	return nil
}

/*
  RebootServer will reboot the server with the `server_id`.

  This function implements the interface described in:-
  * https://docs.hpcloud.com/api/compute/
  * Section 4.4.7.1 Reboot Server
*/
func (a Access) RebootServer(server_id string) error {
	/*
			 The docs mention that a hard reboot will be used
		     no matter what, so there's no point making a type
		     or make the type of reboot an option
	*/
	s := `{"reboot":{"type":"HARD"}}`
	_, err := a.baseComputeRequest(
		fmt.Sprintf("servers/%s/action", server_id),
		"POST", strings.NewReader(s),
	)
	if err != nil {
		return err
	}
	return nil

}

/*
  ListFlavors will list all the available flavours
  on the HPCloud compute API.
*/
func (a Access) ListFlavors() (*Flavors, error) {
	body, err := a.baseComputeRequest("flavors", "GET", nil)
	if err != nil {
		return nil, err
	}

	fl := &Flavors{}
	err = json.Unmarshal(body, fl)
	if err != nil {
		return nil, err
	}
	return fl, nil
}

func (a Access) ListImages() (*Images, error) {
	body, err := a.baseComputeRequest("images", "GET", nil)
	if err != nil {
		return nil, err
	}
	im := &Images{}
	err = json.Unmarshal(body, im)
	if err != nil {
		return nil, err
	}
	return im, nil
}

func (a Access) DeleteImage(image_id string) error {
	_, err := a.baseComputeRequest(
		fmt.Sprintf("images/%s", image_id), "DELETE", nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func (a Access) ListImage(image_id string) (*Image, error) {
	body, err := a.baseComputeRequest(
		fmt.Sprintf("images/%s", image_id), "GET", nil,
	)
	if err != nil {
		return nil, err
	}
	i := &Image{}
	err = json.Unmarshal(body, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

/*
  baseComputeRequest encapsulates the main basic request
  which is done for each endpoint in the Compute API.

  In the ComputeAPI all endpoints generally succeed on
  a 200/202 return code and fail on the usual fail codes.

  We simply check for the known good return codes and return
  the body in those cases or we fail with the appropriate
  response.
*/
func (a Access) baseComputeRequest(url, method string, b io.Reader) ([]byte, error) {
	path := fmt.Sprintf("%s%s/%s", COMPUTE_URL, a.TenantID, url)
	return a.baseRequest(path, method, b)
}

/*
  MarshalJSON implements the Marshaler interface for the
  Server type.

  We implement this interface because when creating a server
  we have optional values and since Go has zero-values and
  does *not* have configurable zero values we need to make
  sure that zero-values are converted to known good values.

  As such:
    * FlavorRef is checked if it's a valid reference.
    * Ditto for ImageRef.
    * Name cannot be blank.
    * If the key is missing, it'll not put anything in.
    * The config_drive defaults to false anyway, no need
      to send a false value.
    * Min/MaxCount are ignored if they are zero.
    * UserData is ignored if it's a blank string.
    * Personality is ignored if it's a blank string.
    * Metadata/SecurityGroups are ignored if they have len(0)
*/
func (s Server) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString("")
	b.WriteString(`{"server":{`)
	/* The available images are 100-105, x-small to x-large. */
	if s.FlavorRef < 100 || s.FlavorRef > 105 {
		return []byte{},
			errors.New("Flavor Reference refers to a non-existant flavour.")
	} else {
		b.WriteString(fmt.Sprintf(`"flavorRef":%d`, s.FlavorRef))
	}
	if s.ImageRef == 0 {
		return []byte{},
			errors.New("An image name is required.")
	} else {
		b.WriteString(fmt.Sprintf(`,"imageRef":%d`, s.ImageRef))
	}
	if s.Name == "" {
		return []byte{},
			errors.New("A name is required")
	} else {
		b.WriteString(fmt.Sprintf(`,"name":"%s"`, s.Name))
	}

	/* Optional items */
	/* The max size of a personality string is 255 bytes. */
	if len(s.Personality) > 255 {
		return []byte{},
			errors.New("Server's personality cannot have >255 bytes.")
	} else if s.Personality != "" {
		b.WriteString(fmt.Sprintf(`,"personality":"%s",`, s.Personality))
	}
	if s.Key != "" {
		b.WriteString(fmt.Sprintf(`,"key_name":"%s"`, s.Key))
	}
	if s.ConfigDrive {
		b.WriteString(`,"config_drive": true`)
	}
	if s.MinCount > 0 {
		b.WriteString(fmt.Sprintf(`,"min_count":%d`, s.MinCount))
	}
	if s.MaxCount > 0 {
		b.WriteString(fmt.Sprintf(`,"max_count":%d`, s.MaxCount))
	}
	if s.UserData != "" {
		/* user_data needs to be base64'd */
		newb := make([]byte, 0, len(s.UserData))
		base64.StdEncoding.Encode([]byte(s.UserData), newb)
		b.WriteString(fmt.Sprintf(`,"user_data": "%s",`, string(newb)))
	}

	/* Ignore the metadata if there isn't any, it's optional. */
	if len(s.Metadata) > 0 {
		fmt.Println(len(s.Metadata))
		b.WriteString(`,"metadata":{`)
		cnt := 0
		for key, value := range s.Metadata {
			b.WriteString(fmt.Sprintf(`"%s": "%s"`, key, value))
			if cnt+1 != len(s.Metadata) {
				b.WriteString(",")
				cnt++
			} else {
				b.WriteString("}")
			}
		}
	}
	/* Ignore the Security Groups if there isn't any, it's optional. */
	if len(s.SecurityGroups) > 0 {
		b.WriteString(`,"security_groups":[`)
		cnt := 0
		for _, sg := range s.SecurityGroups {
			b.WriteString(fmt.Sprintf(`{"name": "%s"}`, sg.Name))
			if cnt+1 != len(s.SecurityGroups) {
				b.WriteString(",")
				cnt++
			} else {
				b.WriteString("]")
			}
		}
	}
	b.WriteString("}}")
	return b.Bytes(), nil
}
