package magicalauth

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// recordObject is a JSON object read by exact member name. encoding/json's
// struct decoding matches names case-insensitively, which would let an
// unchecked "AUD" or "Verified" member override the checked one, so record
// JSON is never decoded straight into a struct. Duplicate exact names: the
// last occurrence wins, as with JSON.parse.
type recordObject map[string]json.RawMessage

func parseRecordObject(raw []byte) (recordObject, bool) {
	var obj recordObject
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return nil, false
	}
	return obj, true
}

// memberReader reads optional members of a recordObject. A present member of
// the wrong JSON type clears ok; an absent member yields the zero value.
type memberReader struct {
	obj recordObject
	ok  bool
}

func newMemberReader(obj recordObject) *memberReader {
	return &memberReader{obj: obj, ok: true}
}

func (r *memberReader) str(name string) string {
	raw, present := r.obj[name]
	if !present {
		return ""
	}
	s, ok := jsonString(raw)
	r.ok = r.ok && ok
	return s
}

func (r *memberReader) boolPtr(name string) *bool {
	raw, present := r.obj[name]
	if !present {
		return nil
	}
	b, ok := jsonBool(raw)
	r.ok = r.ok && ok
	return &b
}

func (r *memberReader) boolean(name string) bool {
	if b := r.boolPtr(name); b != nil {
		return *b
	}
	return false
}

func (r *memberReader) integer(name string) int {
	raw, present := r.obj[name]
	if !present {
		return 0
	}
	n, err := strconv.Atoi(string(bytes.TrimSpace(raw)))
	r.ok = r.ok && err == nil
	return n
}

func (r *memberReader) intOr(name string, def int) int {
	if _, present := r.obj[name]; !present {
		return def
	}
	return r.integer(name)
}

// object returns a nested object member; absent yields nil. A present
// null or other non-object is a wrong JSON type and clears ok.
func (r *memberReader) object(name string) recordObject {
	raw, present := r.obj[name]
	if !present {
		return nil
	}
	obj, ok := parseRecordObject(raw)
	r.ok = r.ok && ok
	return obj
}

func readCarrier(r *memberReader) *VerificationRecordCarrier {
	obj := r.object("carrier")
	if obj == nil {
		return nil
	}
	s := newMemberReader(obj)
	c := &VerificationRecordCarrier{ID: s.str("id"), Name: s.str("name"), MVNO: s.str("mvno")}
	r.ok = r.ok && s.ok && c.ID != ""
	return c
}

func readPhoneNumberDetails(r *memberReader) *PhoneNumberDetails {
	obj := r.object("phone_number_details")
	if obj == nil {
		return nil
	}
	s := newMemberReader(obj)
	d := &PhoneNumberDetails{CountryCode: s.integer("country_code"), NationalNumber: s.str("national_number"), RegionCode: s.str("region_code")}
	r.ok = r.ok && s.ok
	return d
}

func readLiveCall(r *memberReader, name string) *VerificationRecordLiveCall {
	obj := r.object(name)
	if obj == nil {
		return nil
	}
	s := newMemberReader(obj)
	info := &VerificationRecordLiveCall{Status: s.str("status"), CarrierName: s.str("carrier_name"), Reason: s.str("reason")}
	r.ok = r.ok && s.ok && info.Status != ""
	return info
}

func readSwap(r *memberReader, name string) *VerificationRecordSwap {
	obj := r.object(name)
	if obj == nil {
		return nil
	}
	s := newMemberReader(obj)
	info := &VerificationRecordSwap{
		Checked: s.boolean("checked"), RiskLevel: s.str("risk_level"), CarrierName: s.str("carrier_name"),
		Reason: s.str("reason"), Swapped: s.boolPtr("swapped"), MaxAgeHours: s.integer("max_age_hours"),
	}
	if band := s.object("age_band"); band != nil {
		b := newMemberReader(band)
		info.AgeBand = &VerificationRecordAgeBand{From: b.intOr("from", -1), To: b.intOr("to", -1)}
		s.ok = s.ok && b.ok
	}
	r.ok = r.ok && s.ok
	return info
}

// jsonBool accepts only the JSON literals true and false.
func jsonBool(raw json.RawMessage) (bool, bool) {
	switch string(bytes.TrimSpace(raw)) {
	case "true":
		return true, true
	case "false":
		return false, true
	}
	return false, false
}

// jsonInt accepts only a plain JSON integer literal (no fraction, exponent or quotes).
func jsonInt(raw json.RawMessage) (int64, bool) {
	n, err := strconv.ParseInt(string(bytes.TrimSpace(raw)), 10, 64)
	return n, err == nil
}
