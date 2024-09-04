package main

import (
    "bufio"
    "context"
    "io/ioutil"
    "strings"
    "sync"
    "time"

    "github.com/coredns/coredns/plugin"
    "github.com/miekg/dns"
)

type DomainForwarder struct {
    Next             plugin.Handler
    Domains          map[string]struct{}
    PrimaryServer    string
    SecondaryServers []string
    mu               sync.RWMutex
}

func (d *DomainForwarder) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
    if r.MsgHdr.Rcode != dns.RcodeSuccess {
        return dns.RcodeServerFailure, nil
    }

    qName := dns.Fqdn(r.Question[0].Name)
    d.mu.RLock()
    _, found := d.Domains[qName]
    d.mu.RUnlock()

    if found {
        return d.forwardQuery(w, r, d.PrimaryServer)
    }
    return d.forwardQuery(w, r, d.SecondaryServers[0])
}

func (d *DomainForwarder) forwardQuery(w dns.ResponseWriter, r *dns.Msg, server string) (int, error) {
    client := new(dns.Client)
    resp, _, err := client.Exchange(r, server+":53")
    if err != nil {
        return dns.RcodeServerFailure, err
    }
    w.WriteMsg(resp)
    return dns.RcodeSuccess, nil
}

func (d *DomainForwarder) Reload() error {
    d.mu.Lock()
    defer d.mu.Unlock()

    data, err := ioutil.ReadFile("domain.txt")
    if err != nil {
        return err
    }

    domains := make(map[string]struct{})
    scanner := bufio.NewScanner(strings.NewReader(string(data)))
    for scanner.Scan() {
        domain := dns.Fqdn(scanner.Text())
        domains[domain] = struct{}{}
    }

    d.Domains = domains
    return nil
}

func (d *DomainForwarder) Name() string {
    return "domainforwarder"
}

func New(next plugin.Handler) plugin.Handler {
    df := &DomainForwarder{
        Next:             next,
        Domains:          make(map[string]struct{}),
        PrimaryServer:    "202.58.203.196",
        SecondaryServers: []string{"1.1.1.1", "8.8.8.8"},
    }
    // Initial load
    if err := df.Reload(); err != nil {
        panic(err)
    }

    // Reload the file periodically
    go func() {
        for {
            time.Sleep(1 * time.Minute)
            if err := df.Reload(); err != nil {
                // Handle error
                continue
            }
        }
    }()

    return df
}
