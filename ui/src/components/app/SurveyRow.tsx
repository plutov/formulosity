'use client'

import { ErrCode } from 'components/ui/ErrCode'
import { Alert, Badge, Button, Table } from 'flowbite-react'
import { updateSurvey } from 'lib/api'
import { Survey, SurveyDeliveryStatus, SurveyParseStatus } from 'lib/types'
import moment from 'moment'
import { useState } from 'react'
import {
  HiChevronDown,
  HiChevronUp,
  HiExternalLink,
  HiOutlinePause,
  HiOutlinePlay,
} from 'react-icons/hi'

type SurveyCardProps = {
  survey: Survey
}

export function SurveyRow({ survey }: SurveyCardProps) {
  const [errorMsg, setErrorMsg] = useState<string>('')
  const [showErrorLog, setShowErrorLog] = useState<boolean>(false)

  async function updateSurveyStatus(surveyUUID: string, status: string) {
    const res = await updateSurvey(surveyUUID, {
      delivery_status: status,
    })

    if (res.error) {
      setErrorMsg(res.error)
    } else {
      window.location.href = `/app`
    }
  }

  const parseStatusColors = new Map<string, string>([
    [SurveyParseStatus.Success, 'success'],
    [SurveyParseStatus.Error, 'failure'],
    [SurveyParseStatus.Deleted, 'warning'],
  ])
  const isLaunched = survey.delivery_status === SurveyDeliveryStatus.Launched
  const canSartSurvey =
    survey.parse_status === SurveyParseStatus.Success && !isLaunched

  return (
    <Table.Row className="bg-gray-800" key={survey.uuid}>
      <Table.Cell className="bg-gray-800">
        <div>
          <div className="text-base font-semibold">{survey.name}</div>
          {survey.config && (
            <div className="font-normal">{survey.config.title}</div>
          )}
          <div className="font-normal">
            Created on: {moment(survey.created_at).format('MMM D, YYYY')}
          </div>
        </div>
      </Table.Cell>
      <Table.Cell>
        <Badge
          size="xs"
          className="cursor-pointer"
          color={parseStatusColors.get(survey.parse_status)}
          onClick={() => setShowErrorLog(!showErrorLog)}
        >
          {survey.parse_status}&nbsp;
          {survey.parse_status === SurveyParseStatus.Error &&
            (showErrorLog ? (
              <HiChevronUp className="inline" />
            ) : (
              <HiChevronDown className="inline" />
            ))}
        </Badge>
        {showErrorLog && survey.parse_status === SurveyParseStatus.Error && (
          <div>
            <Alert color="dark" rounded>
              <span>
                <p>
                  <span className="font-medium">Error log:</span>
                  <br />
                  <code>{survey.error_log}</code>
                </p>
              </span>
            </Alert>
          </div>
        )}
      </Table.Cell>
      <Table.Cell>
        {(isLaunched || canSartSurvey) && (
          <Button
            className="h-8 bg-crimson-9 enabled:hover:bg-crimson-11 px-2 py-0.5 rounded text-sm flex items-center gap-1 whitespace-nowrap"
            onClick={async () => {
              updateSurveyStatus(
                survey.uuid,
                isLaunched ? 'stopped' : 'launched'
              )
            }}
          >
            {isLaunched ? (
              <>
                <HiOutlinePause className="w-4 h-4" />
                <span>Stop</span>
              </>
            ) : (
              <>
                <HiOutlinePlay className="w-4 h-4" />
                <span>Start</span>
              </>
            )}
          </Button>
        )}
        {errorMsg && (
          <div className="w-full">
            <ErrCode message={errorMsg} />
          </div>
        )}
      </Table.Cell>
      <Table.Cell>
        {survey.delivery_status === SurveyDeliveryStatus.Launched && (
          <a
            href={survey.url}
            target="_blank"
            className="text-crimson-9 hover:text-crimson-11"
          >
            Public Link <HiExternalLink className="inline" />
          </a>
        )}
      </Table.Cell>
      <Table.Cell>
        <a
          href={`/app/surveys/${survey.uuid}/responses`}
          className="text-crimson-9 hover:text-crimson-11"
        >
          {survey.stats.sessions_count_completed}
        </a>
      </Table.Cell>
      <Table.Cell>{survey.stats.completion_rate} %</Table.Cell>
    </Table.Row>
  )
}
