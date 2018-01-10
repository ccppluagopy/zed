import("io/ioutil")

func GetBody() ([]byte, error) {
    buf, err:=ioutil.ReadAll(req.Request.Body)
    return buf, err
}
