'use client'

import { Table } from 'flowbite-react'
import { Survey } from 'lib/types'
import { SurveyRow } from './SurveyRow'

type SurveysPageProps = {
  surveys: Array<Survey>
}

export function SurveysPage({ surveys }: SurveysPageProps) {
  return (
    <div>
      <div className="flex flex-col w-full tabs">
        <div className="flex flex-col w-full items-center">
          <div className="w-full">
            <Table className="text-gray-100">
              <Table.Head>
                <Table.HeadCell>Name/Title</Table.HeadCell>
                <Table.HeadCell>Build</Table.HeadCell>
                <Table.HeadCell>Delivery</Table.HeadCell>
                <Table.HeadCell>Share</Table.HeadCell>
                <Table.HeadCell>Responses</Table.HeadCell>
                <Table.HeadCell>Completion</Table.HeadCell>
              </Table.Head>
              <Table.Body className="divide-y">
                {surveys.map((survey) => {
                  return <SurveyRow key={survey.uuid} survey={survey} />
                })}
              </Table.Body>
            </Table>
          </div>
        </div>
      </div>
    </div>
  )
}
