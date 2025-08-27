const API_BASE_URL = process.env.NEXT_PUBLIC_API_ADDR || 'http://localhost:9900'

export async function call(path: string, init?: RequestInit) {
  try {
    if (init) {
      init['cache'] = 'no-store'
    }

    console.log('calling', `${API_BASE_URL}${path}`)
    const res = await fetch(`${API_BASE_URL}${path}`, init)
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

export async function postFormData(path: string, payload: FormData) {
  return call(path, {
    method: 'POST',
    body: payload,
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

export async function getSurvey(urlSlug: string) {
  const headers = {
    Referer: typeof window !== 'undefined' ? window.location.host : '',
  }

  return await call(`/surveys/${urlSlug}`, {
    method: 'GET',
    headers: headers,
  })
}

export async function getSurveys() {
  return await get(`/app/surveys`)
}

export async function createSurveySession(urlSlug: string) {
  const headers = {
    'Content-Type': 'application/json',
    Referer: typeof window !== 'undefined' ? window.location.host : '',
  }

  return await call(`/surveys/${urlSlug}/sessions`, {
    method: 'PUT',
    body: JSON.stringify({}),
    headers: headers,
  })
}

export async function getSurveySession(urlSlug: string, sessionId: string) {
  const headers = {
    Referer: typeof window !== 'undefined' ? window.location.host : '',
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

export async function deleteSurveySession(
  surveyUUID: string,
  sessionUUID: string
) {
  return await del(`/app/surveys/${surveyUUID}/sessions/${sessionUUID}`)
}

export async function submitQuestionAnswer(
  urlSlug: string,
  sessionId: string,
  questionUUID: string,
  payload: object | FormData
) {
  if (payload instanceof FormData) {
    return await postFormData(
      `/surveys/${urlSlug}/sessions/${sessionId}/questions/${questionUUID}/answers`,
      payload
    )
  } else {
    return await post(
      `/surveys/${urlSlug}/sessions/${sessionId}/questions/${questionUUID}/answers`,
      payload
    )
  }
}

export async function download(surveyUUID: string, fileName: string) {
  const path = `/app/surveys/${surveyUUID}/download/${fileName}`
  const res = await fetch(`${API_BASE_URL}${path}`)
  const blob = await res.blob()
  const fileUrl = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = fileUrl
  a.download = fileName
  a.click()
  URL.revokeObjectURL(fileUrl)
}
