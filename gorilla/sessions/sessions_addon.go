package sessions

import (
	"net/http"
)


// Init creates "__sessionid__" (unique id) and puts it into cookie
func Init(r *http.Request, w http.ResponseWriter) []error {                      
        return DefaultSessionFactory.Init(r, w)
}
 

// Init creates "__sessionid__" (unique id) and puts it into cookie
func (f *SessionFactory) Init(r *http.Request,
        w http.ResponseWriter) []error {
        return getRequestSessions(f, r).Init(w)
}

// Init creates "__sessionid__" (unique id) and puts it into cookie
func (s *requestSessions) Init(w http.ResponseWriter) []error {
        var err error
        var ok bool  
        var errors []error
        for key, info := range s.sessions {
                if ok, err = info.Store.Init(s.request, w, key, &info); !ok {
                        if errors == nil {
                                errors = []error{err}
                        } else {
                                errors = append(errors, err)
                        }
                }
        }
        return errors
}


// Init saves the __sessionid__ in the response.
// CookieSessionStore implementation
func (s *CookieSessionStore) Init(r *http.Request, w http.ResponseWriter,
        key string, info *SessionInfo) (bool, error) {
        return CookieStoreInit(s, w, key, info)
}

// Init saves the __sessionid__ in the response.
// CookieSessionStore implementation
func CookieStoreInit(s SessionStore, w http.ResponseWriter, key string,
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


// GetId helper function returns __sessionid__ as a string
// It can be used in comparison manner for authentication purposes
func (s *SessionData) GetId() string {
        data := SessionData{}
        data = *s
        return data["__sessionid__"].(string)
}
