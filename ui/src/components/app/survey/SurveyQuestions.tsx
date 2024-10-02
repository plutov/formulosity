'use client'

import { useState } from 'react'
import {
  Survey,
  SurveySession,
  SurveyQuestion,
  SurveyQuestionType,
  SurveySessionStatus,
} from 'lib/types'
import {
  Dropdown,
  Button,
  Checkbox,
  TextInput,
  Textarea,
  Datepicker,
  FileInput
} from 'flowbite-react'
import { HiArrowSmRight, HiSelector } from 'react-icons/hi'
import { ErrCode } from 'components/ui/ErrCode'
import { submitQuestionAnswer } from 'lib/api'
import { ReactSortable } from 'react-sortablejs'
import moment from 'moment'
import { DatepicketTheme } from 'components/ui/themes'
import SurveyFooter from './SurveyFooter'
import {
  determineNextQuestion,
  determinePrevQuestionWithAnswer,
  determineNextQuestionWithAnswer,
  determineInitialQuestion,
} from 'lib/questions'

type SurveyQuestionsProps = {
  survey: Survey
  session: SurveySession
  apiURL: string
}

interface SortableItemType {
  id: number
  name: string
}

export default function SurveyQuestions({
  survey,
  session,
  apiURL,
}: SurveyQuestionsProps) {
  const initialQuestion = determineInitialQuestion(survey, session)

  const [currentQuestion, setCurrentQuestion] = useState<
    SurveyQuestion | undefined
  >(initialQuestion)
  const [surveySession, setSurveySession] = useState<SurveySession>(session)

  const [errorMsg, setErrorMsg] = useState<string | undefined>(undefined)
  const [isNextLoading, setIsNextLoading] = useState<boolean>(false)

  // values state
  const [selectedStringValue, setSelectedStringValue] = useState<
    string | undefined
  >(undefined)
  const [selectedArrayValue, setSelectedArrayValue] = useState<string[]>([])
  const [sortableItems, setSortableItems] = useState<SortableItemType[]>([])
  const [selectedFile, setSelectedFile] = useState<File | undefined>(undefined)

  if (surveySession.status === SurveySessionStatus.Completed) {
    return (
      <div className="outro">
        <h1
          className="h2"
          dangerouslySetInnerHTML={{
            __html: survey.config.outro
              ? survey.config.outro.replace(/(?:\r\n|\r|\n)/g, '<br>')
              : 'Response submitted.<br />Thank you!',
          }}
        ></h1>
      </div>
    )
  }

  if (!currentQuestion) {
    return (
      <div className="outro">
        <h1 className="h2">No more questions found in the survey.</h1>

        <div className="flex flex-col py-8 lg:py-16 items-center text-center">
          <Button
            className="try-again"
            onClick={async () => {
              localStorage.removeItem(`survey_session_id:${survey.url_slug}`)
              window.location.reload()
            }}
          >
            Try again
          </Button>
        </div>
      </div>
    )
  }

  const prev = determinePrevQuestionWithAnswer(
    survey,
    surveySession,
    currentQuestion
  )
  const next = determineNextQuestionWithAnswer(
    survey,
    surveySession,
    currentQuestion
  )
  const questionsCount = survey.config.questions.questions.length

  let questionContent = <></>
  switch (currentQuestion.type) {
    case SurveyQuestionType.SingleChoice:
      questionContent = (
        <div className="dropdown">
          <Dropdown
            label={
              selectedStringValue !== undefined
                ? selectedStringValue
                : 'Select one option from the list'
            }
          >
            {currentQuestion.options.map((option, index) => (
              <Dropdown.Item
                key={index}
                onClick={() => {
                  setSelectedStringValue(option)
                }}
              >
                {option}
              </Dropdown.Item>
            ))}
          </Dropdown>
        </div>
      )
      break
    case SurveyQuestionType.MultipleChoice:
      questionContent = (
        <div className="dropdown">
          <Dropdown
            dismissOnClick={false}
            label={
              selectedArrayValue.length > 0
                ? selectedArrayValue.join(', ')
                : 'Select multiple options from the list'
            }
          >
            {currentQuestion.options.map((option, index) => (
              <Dropdown.Item
                key={index}
                onClick={() => {
                  if (selectedArrayValue.includes(option)) {
                    setSelectedArrayValue(
                      selectedArrayValue.filter((v) => v !== option)
                    )
                  } else {
                    setSelectedArrayValue([...selectedArrayValue, option])
                  }
                }}
              >
                <Checkbox
                  className="checkbox"
                  checked={selectedArrayValue.includes(option)}
                />
                &nbsp;
                {option}
              </Dropdown.Item>
            ))}
          </Dropdown>
        </div>
      )
      break
    case SurveyQuestionType.ShortText:
      questionContent = (
        <TextInput
          defaultValue={selectedStringValue || ''}
          placeholder="Type your answer here..."
          required
          onChange={(e) => {
            const newValue = e.target.value === '' ? undefined : e.target.value
            setSelectedStringValue(newValue)
          }}
        />
      )
      break
    case SurveyQuestionType.Email:
      questionContent = (
        <TextInput
          defaultValue={selectedStringValue || ''}
          placeholder="Type your email here..."
          required
          type="email"
          onChange={(e) => {
            const newValue = e.target.value === '' ? undefined : e.target.value
            setSelectedStringValue(newValue)
          }}
        />
      )
      break
    case SurveyQuestionType.LongText:
      questionContent = (
        <Textarea
          defaultValue={selectedStringValue || ''}
          placeholder="Type your answer here..."
          required
          onChange={(e) => {
            const newValue = e.target.value === '' ? undefined : e.target.value
            setSelectedStringValue(newValue)
          }}
        />
      )
      break
    case SurveyQuestionType.Date:
      questionContent = (
        <Datepicker
          className="datepicker"
          theme={DatepicketTheme}
          placeholder="Select the date..."
          defaultDate={
            selectedStringValue ? new Date(selectedStringValue) : new Date()
          }
          value={selectedStringValue || ''}
          showClearButton={false}
          onSelectedDateChanged={(date) => {
            setSelectedStringValue(moment(date).format('YYYY-MM-DD'))
          }}
        />
      )
      break
    case SurveyQuestionType.Rating:
      const numbers = []
      if (currentQuestion.max && currentQuestion.min) {
        for (let i = currentQuestion.min; i <= currentQuestion.max; i++) {
          numbers.push(i)
        }
      }
      questionContent = (
        <Button.Group>
          {numbers.map((number) => {
            const isSelected =
              selectedStringValue !== undefined &&
              number.toString() === selectedStringValue

            return (
              <Button
                key={number}
                className={isSelected ? 'rating-selected' : 'rating'}
                onClick={() => {
                  setSelectedStringValue(number.toString())
                }}
              >
                {number}
              </Button>
            )
          })}
        </Button.Group>
      )
      break
    case SurveyQuestionType.Ranking:
      if (sortableItems.length === 0) {
        setSortableItems(
          currentQuestion.options.map((option, index) => {
            return {
              id: index,
              name: option,
            }
          })
        )
      }

      questionContent = (
        <>
          <p className="caption">Drag and drop items below to rank them.</p>
          <ReactSortable list={sortableItems} setList={setSortableItems}>
            {sortableItems.map((item) => (
              <div key={item.id} className="sortable-item">
                <HiSelector className="inline" />
                &nbsp;
                {item.name}
              </div>
            ))}
          </ReactSortable>
        </>
      )
      break
    case SurveyQuestionType.YesNo:
      questionContent = (
        <Button.Group>
          <Button
            key={'yes'}
            className={
              selectedStringValue === 'yes' ? 'rating-selected' : 'rating'
            }
            onClick={() => {
              setSelectedStringValue('yes')
            }}
          >
            Yes
          </Button>
          <Button
            key={'no'}
            className={
              selectedStringValue === 'no' ? 'rating-selected' : 'rating'
            }
            onClick={() => {
              setSelectedStringValue('no')
            }}
          >
            No
          </Button>
        </Button.Group>
      )
      break
    case SurveyQuestionType.File:
      questionContent = (
        <FileInput
          defaultValue={selectedFile?.name || ''}
          placeholder="Upload Your File here..."
          required
          onChange={(e) => {
            const newValue = e.target.files?.[0] ?? undefined
            setSelectedFile(newValue)
          }}
        />
      )
      break
  }

  async function submitAnswer() {
    if (currentQuestion === undefined) {
      return
    }

    setIsNextLoading(true)

    let payload = {}
    switch (currentQuestion.type) {
      case SurveyQuestionType.SingleChoice:
      case SurveyQuestionType.ShortText:
      case SurveyQuestionType.LongText:
      case SurveyQuestionType.Email:
      case SurveyQuestionType.Date:
        payload = {
          value: selectedStringValue,
        }
        break
      case SurveyQuestionType.MultipleChoice:
        payload = {
          value: selectedArrayValue,
        }
        break
      case SurveyQuestionType.Rating:
        payload = {
          value: parseInt(selectedStringValue as string),
        }
        break
      case SurveyQuestionType.Ranking:
        payload = {
          value: sortableItems.map((item) => item.name),
        }
        break
      case SurveyQuestionType.YesNo:
        payload = {
          value: selectedStringValue === 'yes',
        }
        break
      case SurveyQuestionType.File: {
        const formData = new FormData();
        if (selectedFile) {
          formData.append('file', selectedFile);
        }
        payload = formData;
        break;
      }
    }

    const apiRes = await submitQuestionAnswer(
      survey.url_slug,
      session.uuid,
      currentQuestion.uuid,
      payload,
      apiURL
    )

    if (apiRes.error) {
      setErrorMsg(apiRes.data.error_details || apiRes.error)
    } else {
      const newSession = apiRes.data.data as SurveySession
      setSurveySession(newSession)
      resetValues()

      const nextQuestion = determineNextQuestion(
        survey,
        newSession,
        currentQuestion
      )
      setCurrentQuestion(nextQuestion)
      if (nextQuestion && nextQuestion.answer !== undefined) {
        fillQuestion(nextQuestion)
      }
    }

    setIsNextLoading(false)
  }

  function resetValues() {
    setSelectedStringValue(undefined)
    setSelectedArrayValue([])
    setSortableItems([])
    setErrorMsg(undefined)
  }

  function fillQuestion(question: SurveyQuestion) {
    switch (question.type) {
      case SurveyQuestionType.SingleChoice:
      case SurveyQuestionType.ShortText:
      case SurveyQuestionType.LongText:
      case SurveyQuestionType.Date:
        setSelectedStringValue(question.answer.value as string)
        break
      case SurveyQuestionType.MultipleChoice:
        setSelectedArrayValue(question.answer.value as string[])
        break
      case SurveyQuestionType.Rating:
        const ratingInt = question.answer.value as number
        setSelectedStringValue(ratingInt.toString())
        break
      case SurveyQuestionType.Ranking:
        const rankingArray = question.answer.value as string[]
        setSortableItems(
          rankingArray.map((item, index) => {
            return {
              id: index,
              name: item,
            }
          })
        )
        break
      case SurveyQuestionType.YesNo:
        const yesNoBool = question.answer.value as boolean
        setSelectedStringValue(yesNoBool ? 'yes' : 'no')
        break
      case SurveyQuestionType.File:
    }
  }

  const isSubmitDisabled =
    selectedStringValue === undefined &&
    selectedArrayValue.length === 0 &&
    sortableItems.length === 0 &&
    selectedFile === undefined

  return (
    <>
      <div className="mb-auto flex-grow">
        <div className="mx-auto max-w-screen-md text-center py-8 lg:py-24 gap-4">
          <p className="h4 py-8">{currentQuestion.label}</p>
          {currentQuestion.description && (
            <p className="caption">{currentQuestion.description}</p>
          )}
          <form className="form" onSubmit={async (e) => {
            e.preventDefault()
            await submitAnswer()
          }}>
            {questionContent}</form>
          <div className="w-full flex justify-center mt-8">
            <Button
              className="next-question"
              disabled={isSubmitDisabled}
              isProcessing={isNextLoading}
              onClick={async () => {
                await submitAnswer()
              }}
            >
              Next&nbsp;
              <HiArrowSmRight className="inline" />
            </Button>
          </div>
          {errorMsg && (
            <div className="w-full flex justify-center mt-8">
              <ErrCode message={errorMsg} />
            </div>
          )}
        </div>
      </div>

      <SurveyFooter
        current={currentQuestion.index + 1}
        size={questionsCount}
        canGoBack={prev !== undefined}
        canGoForward={next !== undefined}
        onBackClick={() => {
          if (prev) {
            resetValues()
            setCurrentQuestion(prev)
            fillQuestion(prev)
          }
        }}
        onForwardClick={() => {
          if (next) {
            resetValues()
            setCurrentQuestion(next)
            fillQuestion(next)
          }
        }}
      />
    </>
  )
}
