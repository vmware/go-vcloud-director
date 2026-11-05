* Add transparent retries of transient HTTP failures for all API calls, within the time budget of
  `Client.MaxRetryTimeout`: HTTP 429 and 503 (honouring `Retry-After`), HTTP 500 "request rate
  exceeded" errors, and network errors (except for `POST` requests). A `MaxRetryTimeout` of 0 disables
  the retries [GH-833]
* Add `Client.MaxRetryAttempts` property and `WithMaxRetryAttempts` option to optionally cap the
  total number of attempts of retried requests [GH-833]
