// Copyright (c) 2022 The BFE Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package annotations

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	redirectAnnotationPrefix = BfeAnnotationPrefix + "redirect."

	RedirectURLSetAnnotation       = redirectAnnotationPrefix + "url-set"
	RedirectURLFromQueryAnnotation = redirectAnnotationPrefix + "url-from-query"
	RedirectURLPrefixAddAnnotation = redirectAnnotationPrefix + "url-prefix-add"
	RedirectSchemeSetSetAnnotation = redirectAnnotationPrefix + "scheme-set"

	RedirectResponseStatusAnnotation = redirectAnnotationPrefix + "response-status"
)

const (
	defaultRedirectResponseStatusCode = 302
)

func HasRedirectAnnotations(annotations map[string]string) (bool, error) {
	annotationList := []string{RedirectURLSetAnnotation, RedirectURLFromQueryAnnotation, RedirectURLPrefixAddAnnotation, RedirectSchemeSetSetAnnotation}

	cnt := 0
	for _, annotation := range annotationList {
		if _, ok := annotations[annotation]; ok {
			cnt++
		}
	}

	switch {
	case cnt == 0:
		if annotations[RedirectResponseStatusAnnotation] != "" {
			return false, fmt.Errorf("unexpected annotation: {%s:%s}", RedirectResponseStatusAnnotation, annotations[RedirectResponseStatusAnnotation])
		}
		return false, nil

	case cnt == 1:
		return true, nil

	default:
		return false, errors.New("setting multiple redirection-related annotations at the same time is not supported")
	}
}

// GetRedirectAction try to parse the cmd and the param of the redirection action from the annotations
func GetRedirectAction(annotations map[string]string) (cmd, param string, err error) {
	switch {
	case annotations[RedirectURLSetAnnotation] != "":
		cmd, param = "URL_SET", annotations[RedirectURLSetAnnotation]

	case annotations[RedirectURLFromQueryAnnotation] != "":
		cmd, param = "URL_FROM_QUERY", annotations[RedirectURLFromQueryAnnotation]

	case annotations[RedirectURLPrefixAddAnnotation] != "":
		cmd, param = "URL_PREFIX_ADD", annotations[RedirectURLPrefixAddAnnotation]

	case annotations[RedirectSchemeSetSetAnnotation] != "":
		cmd, param = "SCHEME_SET", annotations[RedirectSchemeSetSetAnnotation]
	}
	return
}

func GetRedirectStatusCode(annotations map[string]string) (int, error) {
	statusCodeStr := annotations[RedirectResponseStatusAnnotation]
	if statusCodeStr == "" {
		return defaultRedirectResponseStatusCode, nil
	}

	statusCodeInt64, err := strconv.ParseInt(statusCodeStr, 10, 64)
	if err != nil || !(statusCodeInt64 >= 300 && statusCodeInt64 <= 399) {
		return 0, fmt.Errorf("the annotation %s should be an integer number with format 3XX", RedirectResponseStatusAnnotation)
	}
	return int(statusCodeInt64), nil
}
