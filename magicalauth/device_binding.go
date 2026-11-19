package magicalauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	// BindingCookiePrefix is the prefix for session-scoped device binding cookies.
	// The full cookie name is BindingCookiePrefix + sessionKey (e.g., "_glide_bind_abc123").
	BindingCookiePrefix = "_glide_bind_"

	// BindingCookieMaxAge is the cookie Max-Age in seconds (5 minutes).
	BindingCookieMaxAge = 300

	cookieExpiresEpoch = "Thu, 01 Jan 1970 00:00:00 GMT"
)

// HexPattern64 matches a 64-character hex string (case-insensitive).
// Exported so callers can reuse it for their own validation.
var HexPattern64 = regexp.MustCompile(`(?i)^[0-9a-f]{64}$`)

const cookieDomainFmt = "; Domain=%s"

// sessionKeyInjectionChars that must not appear in a sessionKey used in cookie names.
var sessionKeyInjectionChars = []string{";", "=", "\r", "\n", " ", "\t"}

// cookieInjectionChars that must not appear in cookie domain or path.
var cookieInjectionChars = []string{";", "\r", "\n"}

// SensitiveBindingKeys for log sanitization. Add these to your logger's redaction list.
var SensitiveBindingKeys = []string{
	"fe_code", "fecode", "fe_hash", "fehash",
	"agg_code", "aggcode", "_glide_bind", "glide_bind",
}

// GenerateFeCode generates a cryptographically random 64-character lowercase hex string
// suitable for use as a frontend binding code.
func GenerateFeCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("magicauth: generate fe_code: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ComputeFeHash computes the SHA-256 hash of an fe_code, returning a 64-character
// lowercase hex string. The input is normalized to lowercase before hashing to ensure
// deterministic output regardless of case.
func ComputeFeHash(feCode string) (string, error) {
	if !IsValidBindingCode(feCode) {
		return "", fmt.Errorf("magicauth: fe_code must be a 64-character hex string")
	}
	normalized := strings.ToLower(feCode)
	h := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(h[:]), nil
}

// IsValidBindingCode returns true if s is a valid 64-character hex string (case-insensitive).
func IsValidBindingCode(s string) bool {
	return HexPattern64.MatchString(s)
}

// GetBindingCookieName returns the full cookie name for a given session key
// (e.g., "_glide_bind_abc123"). Returns an error if sessionKey is invalid.
func GetBindingCookieName(sessionKey string) (string, error) {
	if err := validateSessionKey(sessionKey); err != nil {
		return "", err
	}
	return BindingCookiePrefix + sessionKey, nil
}

// BuildSetBindingCookieHeader returns a complete Set-Cookie header value for the
// session-scoped device binding cookie. The feCode and sessionKey are validated before use.
// The feCode is normalized to lowercase in the cookie value.
func BuildSetBindingCookieHeader(feCode, sessionKey string, opts *BindingCookieOptions) (string, error) {
	if !IsValidBindingCode(feCode) {
		return "", fmt.Errorf("magicauth: fe_code must be a 64-character hex string")
	}

	cookieName, err := GetBindingCookieName(sessionKey)
	if err != nil {
		return "", err
	}

	o := resolveOpts(opts)
	if err := validateCookieOpts(o); err != nil {
		return "", err
	}

	normalizedCode := strings.ToLower(feCode)

	var b strings.Builder
	fmt.Fprintf(&b, "%s=%s", cookieName, normalizedCode)
	b.WriteString("; HttpOnly")
	fmt.Fprintf(&b, "; Max-Age=%d", BindingCookieMaxAge)
	b.WriteString("; SameSite=Lax")
	fmt.Fprintf(&b, "; Path=%s", o.Path)
	if o.Secure {
		b.WriteString("; Secure")
	}
	if o.Domain != "" {
		fmt.Fprintf(&b, cookieDomainFmt, o.Domain)
	}
	return b.String(), nil
}

// BuildClearBindingCookieHeader returns a Set-Cookie header value that expires
// the session-scoped device binding cookie. Includes both Max-Age=0 and Expires
// for broad browser compatibility.
func BuildClearBindingCookieHeader(sessionKey string, opts *BindingCookieOptions) (string, error) {
	cookieName, err := GetBindingCookieName(sessionKey)
	if err != nil {
		return "", err
	}

	o := resolveOpts(opts)
	if err := validateCookieOpts(o); err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s=; HttpOnly; Max-Age=0; Expires=%s; SameSite=Lax; Path=%s",
		cookieName, cookieExpiresEpoch, o.Path)
	if o.Secure {
		b.WriteString("; Secure")
	}
	if o.Domain != "" {
		fmt.Fprintf(&b, cookieDomainFmt, o.Domain)
	}
	return b.String(), nil
}

// ParseBindingCookie extracts the fe_code from a session-scoped _glide_bind_{sessionKey}
// cookie in a raw Cookie header string. The returned value is normalized to lowercase.
// Returns empty string if the cookie is missing or the value is not valid 64-char hex.
func ParseBindingCookie(cookieHeader, sessionKey string) string {
	cookieName, err := GetBindingCookieName(sessionKey)
	if err != nil {
		return ""
	}
	prefix := cookieName + "="
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(part, prefix) {
			continue
		}
		val := strings.TrimPrefix(part, prefix)
		if IsValidBindingCode(val) {
			return strings.ToLower(val)
		}
		return ""
	}
	return ""
}

// ClearStaleBindingCookies returns Set-Cookie header values that expire ALL
// _glide_bind_* cookies found in the given Cookie header. Use this to clean up
// abandoned flows.
func ClearStaleBindingCookies(cookieHeader string, opts *BindingCookieOptions) []string {
	o := resolveOpts(opts)
	if err := validateCookieOpts(o); err != nil {
		return nil
	}

	var headers []string
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		eqIdx := strings.Index(part, "=")
		if eqIdx < 0 {
			continue
		}
		name := part[:eqIdx]
		if !strings.HasPrefix(name, BindingCookiePrefix) {
			continue
		}

		var b strings.Builder
		fmt.Fprintf(&b, "%s=; HttpOnly; Max-Age=0; Expires=%s; SameSite=Lax; Path=%s",
			name, cookieExpiresEpoch, o.Path)
		if o.Secure {
			b.WriteString("; Secure")
		}
		if o.Domain != "" {
			fmt.Fprintf(&b, cookieDomainFmt, o.Domain)
		}
		headers = append(headers, b.String())
	}
	return headers
}

// GetCompletionPageHTML returns the full HTML for the device binding completion
// redirect page. completeEndpoint must be a relative path starting with "/"
// (e.g., "/api/glide/complete"). Cross-origin endpoints are not supported.
//
// On successful completion the page writes a localStorage signal
// (glide_signal_{session_key} = session_key) that the FE SDK in the original
// tab detects via StorageEvent, plus a sibling timestamp
// (glide_sigmeta_{session_key} = epoch ms) the FE SDK uses to sweep stale
// signals. Both keys stay for the session lifetime: the waiting tab may be
// suspended while the App Clip is in the foreground and must still find the
// signal when it resumes. The FE SDK removes them on consume; the page only
// erases them after 10 minutes as a fallback.
//
// Pass CompletionPageOptions{Beacon: true} to also emit same-origin JSON beacons
// (page_start, invalid_link, post_outcome, close, tab_survived) to window.location.pathname.
// Default is off. Beacons never include agg_code. close reports close_outcome
// (unknown | failed) and never close_succeeded.
func GetCompletionPageHTML(completeEndpoint string, opts ...CompletionPageOptions) (string, error) {
	if !strings.HasPrefix(completeEndpoint, "/") || strings.HasPrefix(completeEndpoint, "//") {
		return "", fmt.Errorf(
			"magicauth: completeEndpoint must be a relative path starting with \"/\" (e.g., \"/api/glide/complete\"); " +
				"cross-origin and protocol-relative endpoints are not supported")
	}
	beacon := false
	if len(opts) > 0 {
		beacon = opts[0].Beacon
	}
	return buildCompletionHTML(completeEndpoint, beacon), nil
}

// CompletionPageOptions configures GetCompletionPageHTML.
type CompletionPageOptions struct {
	Beacon bool
}

func validateSessionKey(sessionKey string) error {
	if sessionKey == "" {
		return fmt.Errorf("magicauth: sessionKey must not be empty")
	}
	for _, ch := range sessionKeyInjectionChars {
		if strings.Contains(sessionKey, ch) {
			return fmt.Errorf("magicauth: sessionKey contains invalid character %q", ch)
		}
	}
	return nil
}

func resolveOpts(opts *BindingCookieOptions) BindingCookieOptions {
	if opts == nil {
		return BindingCookieOptions{Path: "/", Secure: true}
	}
	o := *opts
	if o.Path == "" {
		o.Path = "/"
	}
	return o
}

func validateCookieOpts(o BindingCookieOptions) error {
	if !strings.HasPrefix(o.Path, "/") {
		return fmt.Errorf("magicauth: cookie path must start with \"/\"")
	}
	for _, ch := range cookieInjectionChars {
		if strings.Contains(o.Domain, ch) {
			return fmt.Errorf("magicauth: cookie domain contains invalid character %q", ch)
		}
		if strings.Contains(o.Path, ch) {
			return fmt.Errorf("magicauth: cookie path contains invalid character %q", ch)
		}
	}
	return nil
}

// buildCompletionHTML returns the full HTML for the completion redirect page.
// The completeEndpoint is safely interpolated as a JSON string inside the JS
// using json.Marshal to handle all edge cases (control chars, unicode, etc.).
func completionPageScript(safeEndpoint string, beacon bool) string {
	if !beacon {
		return `(async function() {
  var fragment = new URLSearchParams(window.location.hash.substring(1));
  var aggCode = fragment.get('agg_code');
  var sessionKey = fragment.get('session_key');

  window.history.replaceState(null, '', window.location.pathname + window.location.search);

  if (!aggCode || !sessionKey) {
    showError('Invalid verification link', ['This verification link is not valid.', 'Please go back and try again.'], 'link');
    return;
  }

  try {
    var res = await fetch(` + safeEndpoint + `, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ agg_code: aggCode, session_key: sessionKey, user_agent: navigator.userAgent })
    });

    if (res.ok) {
      var signalKey = 'glide_signal_' + sessionKey;
      var sigmetaKey = 'glide_sigmeta_' + sessionKey;
      var signalStored = false;
      try {
        localStorage.setItem(signalKey, sessionKey);
        signalStored = true;
        localStorage.setItem(sigmetaKey, String(Date.now()));
      } catch(e) {}
      if (signalStored) {
        setTimeout(function() {
          try { localStorage.removeItem(signalKey); localStorage.removeItem(sigmetaKey); } catch(e) {}
        }, 600000);
      }
      tryClose();
    } else {
      var code = '';
      try { var data = await res.json(); code = (data && (data.code || data.error)) || ''; } catch (_) {}
      var err = friendlyError(code);
      showError(err.title, err.points, err.illust);
    }
  } catch (e) {
    showError('Connection issue', ['We could not reach the server.', 'Check your internet connection and try again.'], 'wifi');
  }

  function friendlyError(code) {
    switch (code) {
      case 'MISSING_BINDING_COOKIE':
      case 'BROWSER_MISMATCH':
        return { title: 'Open in your default browser', points: ['Carrier verification must be started from your default browser.', 'The redirect will always open in your default browser to complete.'], illust: 'browser' };
      default:
        return { title: 'Verification could not be completed', points: ['Something went wrong during verification.', 'Please go back and try again.'], illust: 'retry' };
    }
  }

  function tryClose() {
    document.querySelector('.spinner').classList.add('hidden');
    window.close();
    try { window.open('', '_self').close(); } catch(e) {}
    setTimeout(function() {
      document.getElementById('hint').textContent = 'You may close this tab.';
      document.getElementById('hint').classList.remove('hidden');
    }, 300);
  }`
	}
	return `(async function() {
  var fragment = new URLSearchParams(window.location.hash.substring(1));
  var aggCode = fragment.get('agg_code');
  var sessionKey = fragment.get('session_key');
  var closeSent = false;
  var survivedSent = false;
  var closeTrigger = '';
  var hiddenAt = 0;

  function pageFields() {
    var o = { has_agg_code: !!aggCode, has_session_key: !!sessionKey };
    if (sessionKey && /^[A-Za-z0-9_-]{1,64}$/.test(sessionKey)) o.session_key = sessionKey;
    return o;
  }

  function sendPageBeacon(event, extra) {
    try {
      var payload = Object.assign({ event: event }, pageFields(), extra || {});
      var url = window.location.pathname;
      var body = JSON.stringify(payload);
      try {
        if (navigator.sendBeacon && navigator.sendBeacon(url, new Blob([body], { type: 'application/json' }))) return;
      } catch (e) {}
      fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body,
        keepalive: true,
        credentials: 'same-origin'
      }).catch(function() {});
    } catch (e) {}
  }

  function showCloseHint() {
    document.getElementById('hint').textContent = 'You may close this tab.';
    document.getElementById('hint').classList.remove('hidden');
  }

  sendPageBeacon('page_start');

  window.history.replaceState(null, '', window.location.pathname + window.location.search);

  if (!aggCode || !sessionKey) {
    sendPageBeacon('invalid_link');
    showError('Invalid verification link', ['This verification link is not valid.', 'Please go back and try again.'], 'link');
    return;
  }

  try {
    var res = await fetch(` + safeEndpoint + `, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ agg_code: aggCode, session_key: sessionKey, user_agent: navigator.userAgent })
    });

    var outcome = { ok: !!res.ok, status: res.status };
    if (res.ok) {
      sendPageBeacon('post_outcome', outcome);
      var signalKey = 'glide_signal_' + sessionKey;
      var sigmetaKey = 'glide_sigmeta_' + sessionKey;
      var signalStored = false;
      try {
        localStorage.setItem(signalKey, sessionKey);
        signalStored = true;
        localStorage.setItem(sigmetaKey, String(Date.now()));
      } catch(e) {}
      if (signalStored) {
        setTimeout(function() {
          try { localStorage.removeItem(signalKey); localStorage.removeItem(sigmetaKey); } catch(e) {}
        }, 600000);
      }
      tryClose();
    } else {
      var code = '';
      try { var data = await res.json(); var raw = data && (data.code || data.error); if (typeof raw === 'string' && /^[A-Za-z0-9_.-]{1,64}$/.test(raw)) { code = raw; outcome.code = raw; } } catch (_) {}
      sendPageBeacon('post_outcome', outcome);
      var err = friendlyError(code);
      showError(err.title, err.points, err.illust);
    }
  } catch (e) {
    sendPageBeacon('post_outcome', { ok: false, status: 0 });
    showError('Connection issue', ['We could not reach the server.', 'Check your internet connection and try again.'], 'wifi');
  }

  function friendlyError(code) {
    switch (code) {
      case 'MISSING_BINDING_COOKIE':
      case 'BROWSER_MISMATCH':
        return { title: 'Open in your default browser', points: ['Carrier verification must be started from your default browser.', 'The redirect will always open in your default browser to complete.'], illust: 'browser' };
      default:
        return { title: 'Verification could not be completed', points: ['Something went wrong during verification.', 'Please go back and try again.'], illust: 'retry' };
    }
  }

  function reportClose(trigger) {
    if (closeSent) return;
    closeSent = true;
    closeTrigger = trigger;
    sendPageBeacon('close', {
      close_outcome: (trigger === 'timeout' && document.visibilityState === 'visible') ? 'failed' : 'unknown',
      trigger: trigger,
      visibility_state: document.visibilityState
    });
  }

  function reportTabSurvived() {
    if (!closeSent || survivedSent) return;
    survivedSent = true;
    var hiddenMs = hiddenAt ? Date.now() - hiddenAt : 0;
    if (hiddenMs < 0) hiddenMs = 0;
    sendPageBeacon('tab_survived', { hidden_ms: hiddenMs, trigger: closeTrigger });
  }

  function tryClose() {
    document.querySelector('.spinner').classList.add('hidden');
    if ((document.hidden || document.visibilityState === 'hidden') && !hiddenAt) hiddenAt = Date.now();
    document.addEventListener('visibilitychange', function() {
      if (document.visibilityState === 'visible') {
        showCloseHint();
        reportTabSurvived();
        hiddenAt = 0;
        return;
      }
      if (!hiddenAt) hiddenAt = Date.now();
      reportClose('visibilitychange');
    });
    window.addEventListener('pagehide', function() {
      if (!hiddenAt) hiddenAt = Date.now();
      reportClose('pagehide');
    });
    window.close();
    try { window.open('', '_self').close(); } catch(e) {}
    setTimeout(function() {
      reportClose('timeout');
      if (document.visibilityState === 'visible') showCloseHint();
    }, 300);
  }`
}

func buildCompletionHTML(completeEndpoint string, beacon bool) string {
	safeBytes, _ := json.Marshal(completeEndpoint)
	safeEndpoint := string(safeBytes) // includes surrounding quotes, e.g. "/api/complete"

	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'">
<title>Verifying</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Text', system-ui, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; min-height: 100dvh; background: #fff; color: #1a1a1a; -webkit-font-smoothing: antialiased; }
  .spinner-wrap { text-align: center; }
  .spinner { width: 44px; height: 44px; border: 3px solid transparent; border-top-color: #222; border-radius: 50%; animation: spin 0.7s linear infinite; margin: 0 auto; }
  @keyframes spin { to { transform: rotate(360deg); } }
  .hidden { display: none; }
  .hint { margin-top: 1.5rem; font-size: 0.875rem; color: #999; }
  .err { display: none; text-align: center; padding: 2rem 1.5rem; max-width: 340px; }
  .err h2 { font-size: 1.25rem; font-weight: 600; color: #1a1a1a; margin-bottom: 1rem; letter-spacing: -0.02em; line-height: 1.3; }
  .err-points { text-align: left; margin: 0 auto 1.5rem; max-width: 280px; }
  .err-points li { font-size: 0.875rem; line-height: 1.6; color: #555; margin-bottom: 0.375rem; padding-left: 0.25rem; }
  .illust { margin: 0 auto 1.5rem; }
  .illust svg { color: #94a3b8; }
  .illust-browser { width: 140px; height: 90px; position: relative; }
  .illust-browser .b { position: absolute; width: 52px; height: 40px; border: 2px solid #cbd5e1; border-radius: 6px; background: #f8fafc; }
  .illust-browser .b .bar { height: 10px; border-bottom: 1.5px solid #e2e8f0; display: flex; align-items: center; padding: 0 4px; gap: 2px; }
  .illust-browser .b .bar i { width: 3px; height: 3px; border-radius: 50%; background: #cbd5e1; }
  .illust-browser .b1 { left: 10px; top: 16px; } .illust-browser .b2 { right: 10px; top: 16px; }
  .illust-browser .aw { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); }
  .illust-browser .aw svg { width: 36px; height: 36px; animation: arrowSpin 2.5s ease-in-out infinite; }
  @keyframes arrowSpin { 0%,100% { transform: rotate(0); } 50% { transform: rotate(180deg); } }
  .illust-retry svg { width: 48px; height: 48px; animation: retryPulse 2s ease-in-out infinite; }
  @keyframes retryPulse { 0%,100% { transform: scale(1); opacity: 0.7; } 50% { transform: scale(1.1); opacity: 1; } }
  .illust-wifi svg { width: 48px; height: 48px; animation: wifiBlink 2s ease-in-out infinite; }
  @keyframes wifiBlink { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
  .illust-link svg { width: 48px; height: 48px; animation: linkShake 2s ease-in-out infinite; }
  @keyframes linkShake { 0%,100% { transform: rotate(0); } 25% { transform: rotate(-5deg); } 75% { transform: rotate(5deg); } }
  .badge { display: inline-flex; align-items: center; background: #fff4e6; border: 1px solid #ffd8a8; border-radius: 100px; padding: 5px 14px; font-size: 0.6875rem; font-weight: 600; color: #e67700; text-transform: uppercase; letter-spacing: 0.06em; margin-top: 0.25rem; }
  @keyframes fadeUp { from { opacity: 0; transform: translateY(12px); } to { opacity: 1; transform: translateY(0); } }
  .err.show { display: block; animation: fadeUp 0.35s ease-out; }
</style>
</head>
<body>
<div class="spinner-wrap" id="loading" role="status" aria-live="polite">
  <div class="spinner"></div>
  <p class="hint hidden" id="hint"></p>
</div>
<div class="err" id="error" role="alert" aria-live="assertive">
  <h2 id="err-title"></h2>
  <ol class="err-points" id="err-points"></ol>
  <div class="illust illust-browser hidden" id="ill-browser">
    <div class="illust-browser">
      <div class="b b1"><div class="bar"><i></i><i></i><i></i></div></div>
      <div class="b b2"><div class="bar"><i></i><i></i><i></i></div></div>
      <div class="aw"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/></svg></div>
    </div>
  </div>
  <div class="illust illust-retry hidden" id="ill-retry">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 11-2.12-9.36L23 10"/></svg>
  </div>
  <div class="illust illust-wifi hidden" id="ill-wifi">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M1 1l22 22M16.72 11.06A10.94 10.94 0 0119 12.55M5 12.55a10.94 10.94 0 015.17-2.39M10.71 5.05A16 16 0 0122.56 9M1.42 9a15.91 15.91 0 014.7-2.88M8.53 16.11a6 6 0 016.95 0M12 20h.01"/></svg>
  </div>
  <div class="illust illust-link hidden" id="ill-link">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M15 7h3a5 5 0 010 10h-3m-6 0H6a5 5 0 010-10h3"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
  </div>
  <div class="badge">Action required</div>
</div>
<script>
` + completionPageScript(safeEndpoint, beacon) + `
  function showError(title, points, illust) {
    document.getElementById('loading').classList.add('hidden');
    document.getElementById('err-title').textContent = title;
    var ol = document.getElementById('err-points');
    ol.innerHTML = '';
    (points || []).forEach(function(p) { var li = document.createElement('li'); li.textContent = p; ol.appendChild(li); });
    var ids = ['ill-browser','ill-retry','ill-wifi','ill-link'];
    ids.forEach(function(id) { document.getElementById(id).classList.add('hidden'); });
    var el = document.getElementById('ill-' + (illust || 'retry'));
    if (el) el.classList.remove('hidden');
    document.getElementById('error').classList.add('show');
  }
})();
</script>
</body>
</html>`
}
