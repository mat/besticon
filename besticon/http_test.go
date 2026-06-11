package besticon

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestGetFollowsRedirectToPrivateHost verifies that a public host which
// redirects to a loopback/private address cannot be used to reach an internal
// service. The initial-host check alone is not enough: the redirect target
// must be re-validated too.
func TestGetFollowsRedirectToPrivateHost(t *testing.T) {
	var internalHit int32

	// "Internal" service on loopback. A direct request to it is rejected by
	// the initial-host check; reaching it requires the redirect bypass.
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&internalHit, 1)
		w.Write([]byte("INTERNAL"))
	}))
	defer internal.Close()

	// Public decoy: bound to the host link-local address (non-loopback,
	// non-private per net.IP), so the initial-host check allows it. It
	// redirects to the loopback internal service.
	llHost := linkLocalHost(t)
	ln, err := net.Listen("tcp", llHost+":0")
	if err != nil {
		t.Skipf("cannot bind link-local %s: %v", llHost, err)
	}
	decoy := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL+"/secret", http.StatusFound)
	})}
	go decoy.Serve(ln)
	defer decoy.Close()

	b := New()
	resp, err := b.Get("http://" + ln.Addr().String() + "/favicon")
	if resp != nil {
		resp.Body.Close()
	}

	if atomic.LoadInt32(&internalHit) != 0 {
		t.Fatalf("redirect bypass: internal loopback service was reached via 302 (hits=%d)", internalHit)
	}
	if err == nil {
		t.Fatalf("expected redirect to private host to be rejected, got nil error")
	}
	if !strings.Contains(err.Error(), "private ip address disallowed") {
		t.Fatalf("expected 'private ip address disallowed', got: %v", err)
	}
}

// TestGetRejectsDirectPrivateHost is the negative control: a direct request to
// a loopback address is already rejected by the initial-host check.
func TestGetRejectsDirectPrivateHost(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("INTERNAL"))
	}))
	defer internal.Close()

	b := New()
	resp, err := b.Get(internal.URL + "/secret")
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "private ip address disallowed") {
		t.Fatalf("expected direct loopback request to be rejected, got err=%v", err)
	}
}

// linkLocalHost returns a usable non-loopback, non-private IP (link-local) on
// the host so the decoy server passes the initial-host check.
func linkLocalHost(t *testing.T) string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		t.Skipf("cannot enumerate interfaces: %v", err)
	}
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.To4() == nil {
			continue
		}
		if ip.IsLinkLocalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() {
			return ip.String()
		}
	}
	t.Skip("no usable link-local IPv4 address on host")
	return ""
}
