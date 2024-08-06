export function getConsoleApiHost() {
  return process.env.NEXT_PUBLIC_CONSOLE_API_ADDR
}

export async function call(path: string, init?: RequestInit) {
  try {
    if (init) {
      init['cache'] = 'no-store'
    }

    const res = await fetch(`${getConsoleApiHost()}${path}`, init)
    const data = await res.json()

    if (!data) {
      return {
        status: 500,
        error: 'invalid response',
        data: {},
      }
    }

    if (res.status !== 200) {
      return {
        status: res.status,
        error: data.message ?? 'unable to call the api',
        data: data,
      }
    }

    return {
      status: 200,
      error: '',
      data: data,
    }
  } catch (e) {
    return {
      status: 500,
      error: 'unable to call the api',
      data: {},
    }
  }
}

export async function post(path: string, payload: object) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(path, {
    method: 'POST',
    body: JSON.stringify(payload),
    headers: headers,
  })
}

export async function put(path: string, payload: object) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(path, {
    method: 'PUT',
    body: JSON.stringify(payload),
    headers: headers,
  })
}

export async function patch(path: string, payload: object) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(path, {
    method: 'PATCH',
    body: JSON.stringify(payload),
    headers: headers,
  })
}

export async function del(path: string) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(path, {
    method: 'DELETE',
    headers: headers,
  })
}

export async function get(path: string) {
  const headers = {}

  return call(path, {
    method: 'GET',
    headers: headers,
  })
}

export async function getSurvey(host: string, urlSlug: string) {
  const headers = {
    Referer: host,
  }

  return await call(`/surveys/${urlSlug}`, {
    method: 'GET',
    headers: headers,
  })
}

export async function getSurveys() {
  return await get(`/app/surveys`)
}

export async function createSurveySession(host: string, urlSlug: string) {
  const headers = {
    'Content-Type': 'application/json',
    Referer: host,
  }

  return await call(`/surveys/${urlSlug}/sessions`, {
    method: 'PUT',
    body: JSON.stringify({}),
    headers: headers,
  })
}

export async function getSurveySession(
  host: string,
  urlSlug: string,
  sessionId: string
) {
  const headers = {
    Referer: host,
  }

  return await call(`/surveys/${urlSlug}/sessions/${sessionId}`, {
    method: 'GET',
    headers: headers,
  })
}

export async function getSurveySessions(surveyUUID: string, filter: string) {
  return await get(`/app/surveys/${surveyUUID}/sessions?${filter}`)
}

export async function updateSurvey(surveyUUID: string, payload: object) {
  return await patch(`/app/surveys/${surveyUUID}`, payload)
}

export async function submitQuestionAnswer(
  urlSlug: string,
  sessionId: string,
  questionUUID: string,
  payload: object
) {
  return await post(
    `/surveys/${urlSlug}/sessions/${sessionId}/questions/${questionUUID}/answers`,
    payload
  )
}
