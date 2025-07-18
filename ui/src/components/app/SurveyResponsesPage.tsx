'use client'

import { useState } from 'react'
import { ErrCode } from 'components/ui/ErrCode'
import { Button, Modal, Pagination, Table } from 'flowbite-react'
import {
  HiClock,
  HiCheck,
  HiOutlineEye,
  HiArrowSmDown,
  HiArrowSmUp,
  HiOutlineDownload,
  HiOutlineTrash,
} from 'react-icons/hi'
import {
  Survey,
  SurveyQuestionType,
  SurveySession,
  SurveySessionStatus,
  SurveySessionsLimit,
} from 'lib/types'
import { getSurveySessions, download, deleteSurveySession } from 'lib/api'
import moment from 'moment'

type SurveyResponsesPageProps = {
  currentSurvey: Survey
  apiURL: string
}

export function SurveyResponsesPage({
  currentSurvey,
  apiURL,
}: SurveyResponsesPageProps) {
  currentSurvey = currentSurvey as Survey

  const [currentPage, setCurrentPage] = useState(1)
  const [sortBy, setSortBy] = useState('created_at')
  const [order, setOrder] = useState('desc')
  const [errorMsg, setErrorMsg] = useState<string>('')
  const [sessions, setSessions] = useState(currentSurvey.sessions)
  const [viewSession, setViewSession] = useState<SurveySession | undefined>(
    undefined
  )
  const [downloading, setDownloading] = useState<boolean>(false)

  const onPageChange = async (page: number) => {
    setCurrentPage(page)

    await fetchResponses(page, sortBy, order)
  }

  const downloadFile = async (path: string) => {
    await download(
      currentSurvey.uuid,
      path.substring(path.lastIndexOf('/') + 1),
      apiURL
    )
  }

  const fetchResponses = async (
    page: number,
    sortBy: string,
    order: string
  ) => {
    setErrorMsg('')

    const limit = SurveySessionsLimit
    const offset = (page - 1) * limit
    const surveySessionsResp = await getSurveySessions(
      currentSurvey.uuid,
      `limit=${limit}&offset=${offset}&sort_by=${sortBy}&order=${order}`,
      apiURL
    )

    if (surveySessionsResp.error) {
      setErrorMsg('Unable to load survey sessions')
    } else {
      setSessions(surveySessionsResp.data.data.sessions)
    }
  }

  const deleteSession = async (session: SurveySession) => {
    setErrorMsg('')

    const deleteSessionResp = await deleteSurveySession(
      currentSurvey.uuid,
      session.uuid,
      apiURL
    )

    if (deleteSessionResp.error) {
      setErrorMsg('Unable to delete survey session')
    } else {
      await fetchResponses(1, sortBy, order)
    }
  }

  const cols = [
    { label: 'Session ID', key: 'uuid' },
    { label: 'Status', key: 'status' },
    { label: 'Started at', key: 'created_at' },
    { label: 'Completed at', key: 'completed_at' },
  ]

  return (
    <div>
      {errorMsg && (
        <div className="w-full my-4">
          <ErrCode message={errorMsg} />
        </div>
      )}

      <div className="flex flex-col w-full">
        <div className="my-4">
          <Button
            className="h-8 bg-crimson-9 enabled:hover:bg-crimson-11 p-2"
            onClick={async () => {
              setDownloading(true)

              const allSessionsResp = await getSurveySessions(
                currentSurvey.uuid,
                `limit=1000000&offset=0&sort_by=created_at&order=desc`,
                apiURL
              )

              setDownloading(false)

              if (allSessionsResp.error) {
                setErrorMsg('Unable to export survey sessions')
                return
              }

              const element = document.createElement('a')
              const file = new Blob(
                [JSON.stringify(allSessionsResp.data.data.sessions)],
                {
                  type: 'application/json',
                }
              )
              element.href = URL.createObjectURL(file)
              element.download = 'survey_responses.json'
              document.body.appendChild(element)
              element.click()
            }}
            disabled={downloading}
          >
            <HiOutlineDownload />
            <p className="px-1">Export responses as JSON</p>
          </Button>
        </div>
        <Table className="text-gray-100">
          <Table.Head>
            {cols.map((col) => (
              <Table.HeadCell
                key={col.key}
                className="cursor-pointer"
                onClick={() => {
                  setSortBy(col.key)
                  let newOrder = 'asc'
                  if (sortBy === col.key) {
                    newOrder = order === 'asc' ? 'desc' : 'asc'
                  }
                  setOrder(newOrder)

                  fetchResponses(1, col.key, newOrder)
                }}
              >
                {col.key === sortBy &&
                  (order === 'asc' ? (
                    <HiArrowSmUp className="inline" />
                  ) : (
                    <HiArrowSmDown className="inline" />
                  ))}{' '}
                {col.label}
              </Table.HeadCell>
            ))}
            <Table.HeadCell>Webhook Status</Table.HeadCell>
            <Table.HeadCell>Actions</Table.HeadCell>
          </Table.Head>
          <Table.Body className="divide-y">
            {sessions.map((session) => (
              <Table.Row className="bg-gray-800" key={session.uuid}>
                <Table.Cell>{session.uuid}</Table.Cell>
                <Table.Cell
                  className={`font-medium ${
                    session.status === SurveySessionStatus.Completed
                      ? 'text-emerald-300'
                      : 'text-yellow-300'
                  }`}
                >
                  {session.status === SurveySessionStatus.Completed && (
                    <>
                      <HiCheck className="inline" />
                      &nbsp;Completed
                    </>
                  )}
                  {session.status === SurveySessionStatus.InProgress && (
                    <>
                      <HiClock className="inline" />
                      &nbsp;In Progress
                    </>
                  )}
                </Table.Cell>
                <Table.Cell>
                  {moment(session.created_at).format('MMM D, YYYY h:mm a')}
                </Table.Cell>
                <Table.Cell>
                  {session.completed_at &&
                    moment(session.completed_at).format('MMM D, YYYY h:mm a')}
                </Table.Cell>
                <Table.Cell>{session.webhookData.statusCode}</Table.Cell>
                <Table.Cell>
                  <div style={{ display: 'flex', gap: '8px' }}>
                    <Button
                      className="h-8 bg-crimson-9 enabled:hover:bg-crimson-11 p-1"
                      onClick={() => setViewSession(session)}
                    >
                      <HiOutlineEye />
                    </Button>
                    <Button
                      className="h-8 bg-crimson-9 enabled:hover:bg-crimson-11 p-1"
                      onClick={() => deleteSession(session)}
                    >
                      <HiOutlineTrash />
                    </Button>
                  </div>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      </div>
      {currentSurvey.pages_count > 1 && (
        <div className="mt-8 flex overflow-x-auto sm:justify-center">
          <Pagination
            layout="pagination"
            currentPage={currentPage}
            totalPages={currentSurvey.pages_count}
            onPageChange={onPageChange}
          />
        </div>
      )}
      <Modal
        show={viewSession !== undefined}
        onClose={() => setViewSession(undefined)}
        size="5xl"
        dismissible={true}
      >
        <Modal.Header>
          {viewSession && `Response: ${viewSession.uuid}`}
        </Modal.Header>
        {viewSession && (
          <Modal.Body>
            <Table className="text-gray-100">
              <Table.Head>
                <Table.HeadCell>Question ID</Table.HeadCell>
                <Table.HeadCell>Question</Table.HeadCell>
                <Table.HeadCell>Response</Table.HeadCell>
              </Table.Head>
              <Table.Body className="divide-y">
                {viewSession.question_answers.map((answer) => {
                  const questions =
                    currentSurvey.config.questions.questions || []
                  const question = questions.find(
                    (q) => q.uuid === answer.question_uuid
                  )

                  let response = ''
                  let isFile = false
                  if (question && answer.answer) {
                    switch (question.type) {
                      case SurveyQuestionType.SingleChoice:
                      case SurveyQuestionType.ShortText:
                      case SurveyQuestionType.LongText:
                      case SurveyQuestionType.Date:
                      case SurveyQuestionType.Email:
                        response = answer.answer.value as string
                        break
                      case SurveyQuestionType.MultipleChoice:
                        response = (answer.answer.value as string[]).join(', ')
                        break
                      case SurveyQuestionType.Rating:
                        response = (answer.answer.value as number).toString()
                        break
                      case SurveyQuestionType.Ranking:
                        response = (answer.answer.value as string[]).join(', ')
                        break
                      case SurveyQuestionType.YesNo:
                        response = (answer.answer.value as boolean)
                          ? 'Yes'
                          : 'No'
                        break
                      case SurveyQuestionType.File:
                        isFile = true
                        response = answer.answer.value as string
                        break
                    }
                  }
                  return (
                    <Table.Row
                      className="bg-gray-800"
                      key={answer.question_uuid}
                    >
                      <Table.Cell>{answer.question_id}</Table.Cell>
                      <Table.Cell>{question && question.label}</Table.Cell>
                      <Table.Cell>
                        {isFile ? (
                          <Button
                            className="h-8 bg-crimson-9 enabled:hover:bg-crimson-11 p-2"
                            onClick={() => downloadFile(response)}
                          >
                            <HiOutlineDownload />
                            <p>Download</p>
                          </Button>
                        ) : (
                          <p>{response}</p>
                        )}
                      </Table.Cell>
                    </Table.Row>
                  )
                })}
              </Table.Body>
            </Table>
          </Modal.Body>
        )}
      </Modal>
    </div>
  )
}
