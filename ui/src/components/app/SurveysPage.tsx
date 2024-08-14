'use client'

import { Tabs, Table } from 'flowbite-react'
import { LuClipboardList } from 'react-icons/lu'
import { Survey } from 'lib/types'
import { SurveyRow } from './SurveyRow'

type SurveysPageProps = {
  surveys: Array<Survey>
  apiURL: string
}

export function SurveysPage({ surveys, apiURL }: SurveysPageProps) {
  return (
    <div>
      <div className="flex flex-col w-full tabs">
        <Tabs.Group style="underline">
          <Tabs.Item
            active
            icon={LuClipboardList}
            title="Surveys"
            className="text-crimson-9"
          >
            <div className="flex flex-col w-full items-center">
              <div className="w-full">
                <Table>
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
                      return (
                        <SurveyRow
                          key={survey.uuid}
                          survey={survey}
                          apiURL={apiURL}
                        />
                      )
                    })}
                  </Table.Body>
                </Table>
              </div>
            </div>
          </Tabs.Item>
        </Tabs.Group>
      </div>
    </div>
  )
}
