export async function call(path: string, init?: RequestInit, host?: string) {
  try {
    if (init) {
      init['cache'] = 'no-store'
    }

    if (!host) {
      // Use environment variable or fallback to default API URL
      host = import.meta.env.VITE_API_URL || 'http://localhost:9900'
      console.log('Using API host:', host)
    }

    console.log('calling', `${host}${path}`)
    const res = await fetch(`${host}${path}`, init)
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
    console.error('unable to call the api', e)
    return {
      status: 500,
      error: 'unable to call the api',
      data: {},
    }
  }
}

export async function post(path: string, payload: object, host?: string) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(
    path,
    {
      method: 'POST',
      body: JSON.stringify(payload),
      headers: headers,
    },
    host
  )
}

export async function postFormData(
  path: string,
  payload: FormData,
  host?: string
) {
  return call(
    path,
    {
      method: 'POST',
      body: payload,
    },
    host
  )
}

export async function put(path: string, payload: object, host?: string) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(
    path,
    {
      method: 'PUT',
      body: JSON.stringify(payload),
      headers: headers,
    },
    host
  )
}

export async function patch(path: string, payload: object, host?: string) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(
    path,
    {
      method: 'PATCH',
      body: JSON.stringify(payload),
      headers: headers,
    },
    host
  )
}

export async function del(path: string, host?: string) {
  const headers = {
    'Content-Type': 'application/json',
  }

  return call(
    path,
    {
      method: 'DELETE',
      headers: headers,
    },
    host
  )
}

export async function get(path: string, host?: string) {
  const headers = {}

  return call(
    path,
    {
      method: 'GET',
      headers: headers,
    },
    host
  )
}

export async function getSurvey(host: string, urlSlug: string) {
  const headers = {
    Referer: host,
  }

  return await call(
    `/surveys/${urlSlug}`,
    {
      method: 'GET',
      headers: headers,
    },
    import.meta.env.VITE_API_URL || 'http://localhost:9900'
  )
}

export async function getSurveys() {
  return await get(
    `/app/surveys`,
    import.meta.env.VITE_API_URL || 'http://localhost:9900'
  )
}

export async function createSurveySession(
  host: string,
  urlSlug: string,
  apiURL: string
) {
  const headers = {
    'Content-Type': 'application/json',
    Referer: host,
  }

  return await call(
    `/surveys/${urlSlug}/sessions`,
    {
      method: 'PUT',
      body: JSON.stringify({}),
      headers: headers,
    },
    apiURL
  )
}

export async function getSurveySession(
  host: string,
  urlSlug: string,
  sessionId: string,
  apiURL: string
) {
  const headers = {
    Referer: host,
  }

  return await call(
    `/surveys/${urlSlug}/sessions/${sessionId}`,
    {
      method: 'GET',
      headers: headers,
    },
    apiURL
  )
}

export async function getSurveySessions(
  surveyUUID: string,
  filter: string,
  apiURL: string
) {
  return await get(`/app/surveys/${surveyUUID}/sessions?${filter}`, apiURL)
}

export async function updateSurvey(
  surveyUUID: string,
  payload: object,
  apiURL: string
) {
  return await patch(`/app/surveys/${surveyUUID}`, payload, apiURL)
}

export async function deleteSurveySession(
  surveyUUID: string,
  sessionUUID: string,
  apiURL: string
) {
  return await del(`/app/surveys/${surveyUUID}/sessions/${sessionUUID}`, apiURL)
}

export async function submitQuestionAnswer(
  urlSlug: string,
  sessionId: string,
  questionUUID: string,
  payload: object | FormData,
  apiURL: string
) {
  if (payload! instanceof FormData) {
    return await postFormData(
      `/surveys/${urlSlug}/sessions/${sessionId}/questions/${questionUUID}/answers`,
      payload,
      apiURL
    )
  } else {
    return await post(
      `/surveys/${urlSlug}/sessions/${sessionId}/questions/${questionUUID}/answers`,
      payload,
      apiURL
    )
  }
}

export async function download(
  surveyUUID: string,
  fileName: string,
  apiURL: string
) {
  const host = apiURL
  const path = `/app/surveys/${surveyUUID}/download/${fileName}`
  const res = await fetch(`${host}${path}`)
  const blob = await res.blob()
  const fileUrl = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = fileUrl
  a.download = fileName
  a.click()
  URL.revokeObjectURL(fileUrl)
}
