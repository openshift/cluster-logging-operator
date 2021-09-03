package legacy

const LegacySecureForwardTemplate = `
{{define "legacySecureForward" -}}
<label @_LEGACY_SECUREFORWARD>
  <match **>
    @type copy
    #include legacy secure-forward.conf
    @include /etc/fluent/configs.d/secure-forward/secure-forward.conf
  </match>
</label>
{{end}}
`

const LegacySyslogForwardTemplate = `
{{define "legacySyslog" -}}
<label @_LEGACY_SYSLOG>
  <match **>
    @type copy
    #include legacy Syslog
    @include /etc/fluent/configs.d/syslog/syslog.conf
  </match>
</label>
{{end}}
`
