package fluentbit

const (

	//ConcatCrioFilter ref https://github.com/fluent/fluent-bit/issues/1316#issuecomment-617445617
	ConcatCrioFilter = `
   local reassemble_state = {}

   function reassemble_cri_logs(tag, timestamp, record)
      -- IMPORTANT: reassemble_key must be unique for each parser stream
      -- otherwise entries from different sources will get mixed up.
      -- Either make sure that your parser tags satisfy this or construct
      -- reassemble_key some other way
      local reassemble_key = tag
      -- if partial line, accumulate
      if record.logtag == 'P' then
         if reassemble_state[reassemble_key] == nil then
           reassemble_state[reassemble_key] = ""
         end
         if record.message ~= nil then
           reassemble_state[reassemble_key] = reassemble_state[reassemble_key] .. record.message
         end
         return -1, 0, 0
      end
      -- otherwise it's a full line, concatenate with accumulated partial lines if any
         if reassemble_state[reassemble_key] == nil then
           reassemble_state[reassemble_key] = ""
         end
      record.message = reassemble_state[reassemble_key] .. (record.message or "")
      reassemble_state[reassemble_key] = nil
      return 1, timestamp, record
   end
`
)
