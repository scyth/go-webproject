package sessions

import (
	"net/http"
	"errors"
	"os"
	"io"
//	"fmt"
	"sync"
)

var (
	fileMutex sync.RWMutex
        ErrDecodingFileRead = errors.New("The session value could not be decoded. Read error from file")
        ErrSaveSession      = errors.New("Session could not be saved. Missing __sessionid__.")
	ErrNoInit	    = errors.New("Session has not been initialized with Init()")

)


type FileSessionStore struct {
	encoders []SessionEncoder
}


func (s *FileSessionStore) Load(r *http.Request, key string,
        info *SessionInfo) {
        info.Data = FileStoreGetCookie(s, r, key, info)
}


// Save saves the session in the response.
func (s *FileSessionStore) Save(r *http.Request, w http.ResponseWriter,
        key string, info *SessionInfo) (bool, error) {
        return FileStoreSetCookie(s, w, key, info)
}

// Encoders returns the encoders for this store.
func (s *FileSessionStore) Encoders() []SessionEncoder {
        return s.encoders
}
 

// SetEncoders sets a group of encoders in the store.
func (s *FileSessionStore) SetEncoders(encoders ...SessionEncoder) {
        s.encoders = encoders
}


// Init creates "__sessionid__" (unique id) and puts it into cookie
func (s *FileSessionStore) Init(r *http.Request, w http.ResponseWriter,
        key string, info *SessionInfo) (bool, error) {
        return FileStoreInit(s, w, key, info)
}



// FileStoreGetCookie returns the contents from a session cookie.
//
// If the session is invalid, it will return an empty SessionData.
func FileStoreGetCookie(s SessionStore, r *http.Request, key string, info *SessionInfo) SessionData {
        cookie, err := r.Cookie(key)
	if err == nil {
                if data, err2 := FileStoreDecode(s, key, cookie.Value, r, info); err2 == nil {
			return data
                }
        }
        return SessionData{}
}
 
// FileStoreSetCookie sets a session cookie using the user-defined configuration.
//
// Custom backends will only store a session id in the cookie.
func FileStoreSetCookie(s SessionStore, w http.ResponseWriter, key string,
        info *SessionInfo) (bool, error) {
	// we need only __sessionid__ from info.Data
	var s_id string
	if sid,ok := info.Data["__sessionid__"]; ok {
		s_id = sid.(string)
	} else {
		return false, ErrNoInit
	}

	userData := SessionData{}
	userData["__sessionid__"] = s_id
	 
	
        encoded, err := FileStoreEncode(s, key, userData)  

        if err != nil {
                return false, err
        }
        cookie := &http.Cookie{
                Name:     key, 
                Value:    encoded,
                Path:     info.Config.Path,
                Domain:   info.Config.Domain,
                MaxAge:   info.Config.MaxAge,
                Secure:   info.Config.Secure,
                HttpOnly: info.Config.HttpOnly,
        }
        http.SetCookie(w, cookie)

	// if we have more than just __sessionid__, we'll write it to file
	if len(info.Data) > 1 {
		encoded, err = FileStoreEncode(s, key, info.Data)
		if err != nil {
			return false, err
		}

		fileMutex.Lock()
		fp, err := os.OpenFile("/tmp/sess_gwp"+string(s_id), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fileMutex.Unlock()
			return false, err
		}
		_, err = fp.Write([]byte(encoded))
		fileMutex.Unlock()
		fp.Close()
	}


        return true, nil
}
 
// Encode encodes a session value for a session store.
func FileStoreEncode(s SessionStore, key string, value SessionData) (string, error) {
        encoders := s.Encoders()
        if encoders != nil {
                var encoded string
                var err error
                for _, encoder := range encoders {
                        encoded, err = encoder.Encode(key, value)
                        if err == nil {
                                return encoded, nil
			}
                }
        }
        return "", ErrEncoding
}
 
// FileStoreDecode decodes a session value for a session store.
func FileStoreDecode(s SessionStore, key, value string, r *http.Request, info *SessionInfo) (SessionData, error) {
        encoders := s.Encoders()
        if encoders != nil {
                var decoded SessionData
                var err error
                for _, encoder := range encoders {
                        decoded, err = encoder.Decode(key, value)
			if err == nil {
				if _,ok := decoded["__sessionid__"]; ok==false {
					return nil, ErrNoInit
				}

				// load other data from file
				if fp,e := os.OpenFile("/tmp/sess_gwp"+decoded.GetId(), os.O_RDONLY, 0400); e == nil {
					defer fp.Close()
					var fdata []byte
					buf := make([]byte, 128)
					for {
						n, err := fp.Read(buf[0:])
						fdata = append(fdata, buf[0:n]...)
						if err != nil {
							if err == io.EOF {
								break
							}
							return decoded, nil
						}
					}
					var fdecoded SessionData
					fdecoded, err = encoder.Decode(key, string(fdata))
					if err == nil {
						return fdecoded, nil
					} else { 
						return nil, ErrDecoding
					}
				} else {
					return decoded, nil
				}
			}
                }
        }
        return nil, ErrDecoding
}


// Init saves the __sessionid__ in the response.
// FileStore implementation
func FileStoreInit(s SessionStore, w http.ResponseWriter, key string,
        info *SessionInfo) (bool, error) {
	
	if _,ok := info.Data["__sessionid__"]; ok {
		return true,nil
	} else {
		info.Data["__sessionid__"],_ = GenerateSessionId(16)
	}

        encoded, err := FileStoreEncode(s, key, info.Data)
        if err != nil {
                return false, err
        }

	cookie := &http.Cookie{   
                Name:     key,
                Value:    encoded,
                Path:     info.Config.Path,
                Domain:   info.Config.Domain,
                MaxAge:   info.Config.MaxAge,
                Secure:   info.Config.Secure,
                HttpOnly: info.Config.HttpOnly,
        }
        http.SetCookie(w, cookie)
	return true,nil
}

