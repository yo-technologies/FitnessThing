/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface GetWorkoutsResponseWorkoutDetails {
  workout?: WorkoutWorkout;
  exerciseLogs?: WorkoutExerciseLog[];
}

export interface RoutineServiceAddExerciseToRoutineBody {
  exerciseId: string;
}

export interface RoutineServiceAddSetToExerciseInstanceBody {
  setType: WorkoutSetType;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
}

export interface RoutineServiceSetExerciseOrderBody {
  exerciseInstanceIds?: string[];
}

export interface RoutineServiceUpdateRoutineBody {
  name?: string;
  description?: string;
}

export interface RoutineServiceUpdateSetInExerciseInstanceBody {
  setType: WorkoutSetType;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
}

export interface WorkoutReportResponseAdditionalInfo {
  /** @format int32 */
  totalSets?: number;
  /** @format int32 */
  totalReps?: number;
  /** @format float */
  totalWeight?: number;
  totalTime?: string;
}

export interface WorkoutServiceAddCommentToWorkoutBody {
  comment?: string;
}

export interface WorkoutServiceAddNotesToExerciseLogBody {
  notes?: string;
}

export interface WorkoutServiceAddPowerRatingToExerciseLogBody {
  /** @format int32 */
  powerRating?: number;
}

export type WorkoutServiceCompleteWorkoutBody = object;

export interface WorkoutServiceGenerateWorkoutBody {
  /** пользовательский промпт при перезапуске */
  userPrompt?: string;
}

export interface WorkoutServiceLogExerciseBody {
  exerciseId: string;
}

export interface WorkoutServiceLogSetBody {
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
}

export interface WorkoutServiceRateWorkoutBody {
  /** @format int32 */
  rating?: number;
}

export interface WorkoutServiceUpdateExerciseLogWeightUnitBody {
  weightUnit: WorkoutWeightUnit;
}

export interface WorkoutServiceUpdateSetLogBody {
  setType?: WorkoutSetType;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
}

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: any;
}

export interface RpcStatus {
  /** @format int32 */
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

export interface WorkoutCreateExerciseRequest {
  name: string;
  description?: string;
  videoUrl?: string;
  targetMuscleGroupIds?: string[];
}

export interface WorkoutCreateRoutineRequest {
  workoutId?: string;
  name: string;
  description?: string;
}

export interface WorkoutExercise {
  id?: string;
  /** @format date-time */
  createdAt?: string;
  name?: string;
  description?: string;
  videoUrl?: string;
  targetMuscleGroups?: string[];
  /** @format date-time */
  updatedAt?: string;
}

export interface WorkoutExerciseHistoryResponse {
  exerciseLogs?: WorkoutExerciseLogDetails[];
}

/** Структура подхода */
export interface WorkoutExerciseInstance {
  id?: string;
  exerciseId?: string;
  /** @format date-time */
  createdAt?: string;
  routineId?: string;
  /** @format date-time */
  updatedAt?: string;
}

export interface WorkoutExerciseInstanceDetails {
  exerciseInstance?: WorkoutExerciseInstance;
  exercise?: WorkoutExercise;
  sets?: WorkoutSet[];
}

/** Лог выполнения упражнения */
export interface WorkoutExerciseLog {
  id?: string;
  /** @format date-time */
  createdAt?: string;
  workoutId?: string;
  exerciseId?: string;
  notes?: string;
  /** @format int32 */
  powerRating?: number;
  /** @format date-time */
  updatedAt?: string;
  weightUnit?: WorkoutWeightUnit;
}

export interface WorkoutExerciseLogDetails {
  exerciseLog?: WorkoutExerciseLog;
  exercise?: WorkoutExercise;
  setLogs?: WorkoutSetLog[];
  expectedSets?: WorkoutExpectedSet[];
}

export interface WorkoutExerciseLogResponse {
  exerciseLogDetails?: WorkoutExerciseLogDetails;
}

export interface WorkoutExerciseResponse {
  exercise?: WorkoutExercise;
}

/** Ожидаемый сет */
export interface WorkoutExpectedSet {
  id?: string;
  exerciseLogId?: string;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
}

/** @default "EXPERIENCE_LEVEL_UNSPECIFIED" */
export enum WorkoutExperienceLevel {
  EXPERIENCE_LEVEL_UNSPECIFIED = "EXPERIENCE_LEVEL_UNSPECIFIED",
  EXPERIENCE_LEVEL_BEGINNER = "EXPERIENCE_LEVEL_BEGINNER",
  EXPERIENCE_LEVEL_INTERMEDIATE = "EXPERIENCE_LEVEL_INTERMEDIATE",
  EXPERIENCE_LEVEL_ADVANCED = "EXPERIENCE_LEVEL_ADVANCED",
}

/**
 * Статусы генерации тренировки
 * - GENERATION_STATUS_UNSPECIFIED: не запускалась
 *  - GENERATION_STATUS_RUNNING: в процессе
 *  - GENERATION_STATUS_FAILED: завершилась с ошибкой
 *  - GENERATION_STATUS_COMPLETED: успешно завершена
 * @default "GENERATION_STATUS_UNSPECIFIED"
 */
export enum WorkoutGenerationStatus {
  GENERATION_STATUS_UNSPECIFIED = "GENERATION_STATUS_UNSPECIFIED",
  GENERATION_STATUS_RUNNING = "GENERATION_STATUS_RUNNING",
  GENERATION_STATUS_FAILED = "GENERATION_STATUS_FAILED",
  GENERATION_STATUS_COMPLETED = "GENERATION_STATUS_COMPLETED",
}

export interface WorkoutGetExerciseAlternativesResponse {
  alternatives?: WorkoutExercise[];
}

export interface WorkoutGetExerciseInstanceDetailsResponse {
  exerciseInstanceDetails?: WorkoutExerciseInstanceDetails;
}

export interface WorkoutGetExercisesResponse {
  exercises?: WorkoutExercise[];
}

export interface WorkoutGetMuscleGroupsResponse {
  muscleGroups?: WorkoutMuscleGroup[];
}

export interface WorkoutGetWorkoutResponse {
  workout?: WorkoutWorkout;
  exerciseLogs?: WorkoutExerciseLogDetails[];
}

export interface WorkoutGetWorkoutsResponse {
  workouts?: GetWorkoutsResponseWorkoutDetails[];
}

/** @default "GOAL_UNSPECIFIED" */
export enum WorkoutGoal {
  GOAL_UNSPECIFIED = "GOAL_UNSPECIFIED",
  GOAL_MUSCLE_GAIN = "GOAL_MUSCLE_GAIN",
  GOAL_WEIGHT_LOSS = "GOAL_WEIGHT_LOSS",
  GOAL_STRENGTH = "GOAL_STRENGTH",
  GOAL_ENDURANCE = "GOAL_ENDURANCE",
  GOAL_FLEXIBILITY = "GOAL_FLEXIBILITY",
}

export interface WorkoutMuscleGroup {
  id?: string;
  name?: string;
}

export interface WorkoutPresignUploadRequest {
  filename: string;
  contentType: string;
}

export interface WorkoutPresignUploadResponse {
  uploadUrl?: string;
  getUrl?: string;
}

/** Структура плана тренировки */
export interface WorkoutRoutine {
  id?: string;
  userId?: string;
  name?: string;
  description?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
  /** @format int32 */
  exerciseCount?: number;
}

export interface WorkoutRoutineDetailResponse {
  routine?: WorkoutRoutine;
  exerciseInstances?: WorkoutExerciseInstanceDetails[];
}

export interface WorkoutRoutineInstanceResponse {
  exerciseInstance?: WorkoutExerciseInstance;
}

export interface WorkoutRoutineListResponse {
  routines?: WorkoutRoutine[];
}

export interface WorkoutRoutineResponse {
  routine?: WorkoutRoutine;
}

/** Структура сета (подхода) */
export interface WorkoutSet {
  id?: string;
  /** @format date-time */
  createdAt?: string;
  exerciseInstanceId?: string;
  setType?: WorkoutSetType;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
  /** @format date-time */
  updatedAt?: string;
}

/** Лог выполнения подхода */
export interface WorkoutSetLog {
  id?: string;
  /** @format date-time */
  createdAt?: string;
  exerciseLogId?: string;
  /** @format int32 */
  reps?: number;
  /** @format float */
  weight?: number;
  time?: string;
  /** @format date-time */
  updatedAt?: string;
}

export interface WorkoutSetLogResponse {
  setLog?: WorkoutSetLog;
}

export interface WorkoutSetResponse {
  set?: WorkoutSet;
}

/**
 * Перечень типов подходов
 * @default "SET_TYPE_UNSPECIFIED"
 */
export enum WorkoutSetType {
  SET_TYPE_UNSPECIFIED = "SET_TYPE_UNSPECIFIED",
  SET_TYPE_REPS = "SET_TYPE_REPS",
  SET_TYPE_WEIGHT = "SET_TYPE_WEIGHT",
  SET_TYPE_TIME = "SET_TYPE_TIME",
}

export interface WorkoutStartWorkoutRequest {
  routineId?: string;
  /** deprecated: используйте отдельный эндпоинт GenerateWorkout */
  generateWorkout?: boolean;
  userPrompt?: string;
}

export interface WorkoutUpdateUserRequest {
  /** @format date-time */
  dateOfBirth?: string;
  /** @format float */
  height?: number;
  /** @format float */
  weight?: number;
}

export interface WorkoutUpdateWorkoutGenerationSettingsRequest {
  basePrompt?: string;
  /** @format int32 */
  varietyLevel?: number;
  primaryGoal?: WorkoutGoal;
  secondaryGoals?: string[];
  experienceLevel?: WorkoutExperienceLevel;
  /** @format int32 */
  daysPerWeek?: number;
  /** @format int32 */
  sessionDurationMinutes?: number;
  injuries?: string;
  priorityMuscleGroupsIds?: string[];
  workoutPlanType?: WorkoutWorkoutPlanType;
}

export interface WorkoutUser {
  id?: string;
  /** @format int32 */
  telegramId?: number;
  username?: string;
  firstName?: string;
  lastName?: string;
  profilePictureUrl?: string;
  /** @format float */
  weight?: number;
  /** @format float */
  height?: number;
  /** @format date-time */
  dateOfBirth?: string;
  /** @format date-time */
  createdAt?: string;
  /** @format date-time */
  updatedAt?: string;
  hasCompletedOnboarding?: boolean;
}

export interface WorkoutUserResponse {
  user?: WorkoutUser;
}

/**
 * Единицы измерения веса
 * @default "WEIGHT_UNIT_UNSPECIFIED"
 */
export enum WorkoutWeightUnit {
  WEIGHT_UNIT_UNSPECIFIED = "WEIGHT_UNIT_UNSPECIFIED",
  WEIGHT_UNIT_KG = "WEIGHT_UNIT_KG",
  WEIGHT_UNIT_LB = "WEIGHT_UNIT_LB",
}

/** Структура тренировки */
export interface WorkoutWorkout {
  id?: string;
  /** @format date-time */
  createdAt?: string;
  userId?: string;
  routineId?: string;
  notes?: string;
  /** @format int32 */
  rating?: number;
  /** @format date-time */
  finishedAt?: string;
  /** @format date-time */
  updatedAt?: string;
  isAiGenerated?: boolean;
  reasoning?: string;
  /** Текущий статус генерации тренировки. */
  generationStatus?: WorkoutGenerationStatus;
}

/** Настройки генерации тренировок */
export interface WorkoutWorkoutGenerationSettings {
  basePrompt?: string;
  /** @format int32 */
  varietyLevel?: number;
  primaryGoal?: WorkoutGoal;
  secondaryGoals?: string[];
  experienceLevel?: WorkoutExperienceLevel;
  /** @format int32 */
  daysPerWeek?: number;
  /** @format int32 */
  sessionDurationMinutes?: number;
  injuries?: string;
  priorityMuscleGroupsIds?: string[];
  workoutPlanType?: WorkoutWorkoutPlanType;
  /** @format date-time */
  updatedAt?: string;
}

export interface WorkoutWorkoutGenerationSettingsResponse {
  settings?: WorkoutWorkoutGenerationSettings;
}

/** @default "WORKOUT_PLAN_TYPE_UNSPECIFIED" */
export enum WorkoutWorkoutPlanType {
  WORKOUT_PLAN_TYPE_UNSPECIFIED = "WORKOUT_PLAN_TYPE_UNSPECIFIED",
  WORKOUT_PLAN_TYPE_FULL_BODY = "WORKOUT_PLAN_TYPE_FULL_BODY",
  WORKOUT_PLAN_TYPE_SPLIT = "WORKOUT_PLAN_TYPE_SPLIT",
  WORKOUT_PLAN_TYPE_UPPER_LOWER = "WORKOUT_PLAN_TYPE_UPPER_LOWER",
  WORKOUT_PLAN_TYPE_PUSH_PULL_LEGS = "WORKOUT_PLAN_TYPE_PUSH_PULL_LEGS",
}

export interface WorkoutWorkoutReportResponse {
  workout?: WorkoutWorkout;
  exerciseLogs?: WorkoutExerciseLog[];
  additionalInfo?: WorkoutReportResponseAdditionalInfo;
}

export interface WorkoutWorkoutResponse {
  workout?: WorkoutWorkout;
}

export interface WorkoutWorkoutsListResponse {
  workouts?: WorkoutWorkout[];
}

import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, HeadersDefaults, ResponseType } from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<FullRequestParams, "body" | "method" | "query" | "path">;

export interface ApiConfig<SecurityDataType = unknown> extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({ securityWorker, secure, format, ...axiosConfig }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({ ...axiosConfig, baseURL: axiosConfig.baseURL || "/api" });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(params1: AxiosRequestConfig, params2?: AxiosRequestConfig): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method && this.instance.defaults.headers[method.toLowerCase() as keyof HeadersDefaults]) || {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] = property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(key, isFileType ? formItem : this.stringifyFormItem(formItem));
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (type === ContentType.FormData && body && body !== null && typeof body === "object") {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (type === ContentType.Text && body && body !== null && typeof body !== "string") {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type && type !== ContentType.FormData ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title Fitness Trainer API
 * @version 1.0.0
 * @baseUrl /api
 *
 * API for fitness Trainer service
 */
export class Api<SecurityDataType extends unknown> extends HttpClient<SecurityDataType> {
  v1 = {
    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceGetExercises
     * @summary Метод для получения списка всех упражнений
     * @request GET:/v1/exercises
     * @secure
     */
    exerciseServiceGetExercises: (
      query?: {
        muscleGroupIds?: string[];
        excludeExerciseIds?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<WorkoutGetExercisesResponse, RpcStatus>({
        path: `/v1/exercises`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceCreateExercise
     * @summary Метод для создания нового упражнения
     * @request POST:/v1/exercises
     * @secure
     */
    exerciseServiceCreateExercise: (body: WorkoutCreateExerciseRequest, params: RequestParams = {}) =>
      this.request<WorkoutExerciseResponse, RpcStatus>({
        path: `/v1/exercises`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceGetExerciseDetail
     * @summary Метод для получения деталей об упражнении
     * @request GET:/v1/exercises/{exerciseId}
     * @secure
     */
    exerciseServiceGetExerciseDetail: (exerciseId: string, params: RequestParams = {}) =>
      this.request<WorkoutExerciseResponse, RpcStatus>({
        path: `/v1/exercises/${exerciseId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceGetExerciseAlternatives
     * @summary Метод для получения альтернативных упражнений по exercise_id
     * @request GET:/v1/exercises/{exerciseId}/alternatives
     * @secure
     */
    exerciseServiceGetExerciseAlternatives: (exerciseId: string, params: RequestParams = {}) =>
      this.request<WorkoutGetExerciseAlternativesResponse, RpcStatus>({
        path: `/v1/exercises/${exerciseId}/alternatives`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceGetExerciseHistory
     * @summary Метод для получения истории выполнения упражнения
     * @request GET:/v1/exercises/{exerciseId}/history
     * @secure
     */
    exerciseServiceGetExerciseHistory: (
      exerciseId: string,
      query?: {
        /** @format int32 */
        offset?: number;
        /** @format int32 */
        limit?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<WorkoutExerciseHistoryResponse, RpcStatus>({
        path: `/v1/exercises/${exerciseId}/history`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags FileService
     * @name FileServicePresignUpload
     * @request POST:/v1/files/presign
     * @secure
     */
    fileServicePresignUpload: (body: WorkoutPresignUploadRequest, params: RequestParams = {}) =>
      this.request<WorkoutPresignUploadResponse, RpcStatus>({
        path: `/v1/files/presign`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags ExerciseService
     * @name ExerciseServiceGetMuscleGroups
     * @summary Метод для получения списка групп мышц
     * @request GET:/v1/muscle_groups
     * @secure
     */
    exerciseServiceGetMuscleGroups: (params: RequestParams = {}) =>
      this.request<WorkoutGetMuscleGroupsResponse, RpcStatus>({
        path: `/v1/muscle_groups`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceGetRoutines
     * @summary Метод для получения списка доступных пользователю рутин
     * @request GET:/v1/routines
     * @secure
     */
    routineServiceGetRoutines: (params: RequestParams = {}) =>
      this.request<WorkoutRoutineListResponse, RpcStatus>({
        path: `/v1/routines`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceCreateRoutine
     * @summary Создание новой рутины
     * @request POST:/v1/routines
     * @secure
     */
    routineServiceCreateRoutine: (body: WorkoutCreateRoutineRequest, params: RequestParams = {}) =>
      this.request<WorkoutRoutineResponse, RpcStatus>({
        path: `/v1/routines`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceGetRoutineDetail
     * @summary Получение информации о рутине по ID
     * @request GET:/v1/routines/{routineId}
     * @secure
     */
    routineServiceGetRoutineDetail: (routineId: string, params: RequestParams = {}) =>
      this.request<WorkoutRoutineDetailResponse, RpcStatus>({
        path: `/v1/routines/${routineId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceDeleteRoutine
     * @summary Удаление рутины по ID
     * @request DELETE:/v1/routines/{routineId}
     * @secure
     */
    routineServiceDeleteRoutine: (routineId: string, params: RequestParams = {}) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/routines/${routineId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceUpdateRoutine
     * @summary Обновление рутины по ID
     * @request PUT:/v1/routines/{routineId}
     * @secure
     */
    routineServiceUpdateRoutine: (
      routineId: string,
      body: RoutineServiceUpdateRoutineBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutRoutineResponse, RpcStatus>({
        path: `/v1/routines/${routineId}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceSetExerciseOrder
     * @summary Метод для установки порядка упражнений в рутине
     * @request POST:/v1/routines/{routineId}/exercise_instances/order
     * @secure
     */
    routineServiceSetExerciseOrder: (
      routineId: string,
      body: RoutineServiceSetExerciseOrderBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/order`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceGetExerciseInstanceDetails
     * @summary Получить информацию об упражнении в рутине
     * @request GET:/v1/routines/{routineId}/exercise_instances/{exerciseInstanceId}
     * @secure
     */
    routineServiceGetExerciseInstanceDetails: (
      routineId: string,
      exerciseInstanceId: string,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutGetExerciseInstanceDetailsResponse, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/${exerciseInstanceId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceRemoveExerciseInstanceFromRoutine
     * @summary Удаление упражнения из рутины
     * @request DELETE:/v1/routines/{routineId}/exercise_instances/{exerciseInstanceId}
     * @secure
     */
    routineServiceRemoveExerciseInstanceFromRoutine: (
      routineId: string,
      exerciseInstanceId: string,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/${exerciseInstanceId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceAddSetToExerciseInstance
     * @summary Метод для добавления сета в упражнение
     * @request POST:/v1/routines/{routineId}/exercise_instances/{exerciseInstanceId}/sets
     * @secure
     */
    routineServiceAddSetToExerciseInstance: (
      routineId: string,
      exerciseInstanceId: string,
      body: RoutineServiceAddSetToExerciseInstanceBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutSetResponse, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/${exerciseInstanceId}/sets`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceRemoveSetFromExerciseInstance
     * @summary Метод для удаления сета из упражнения
     * @request DELETE:/v1/routines/{routineId}/exercise_instances/{exerciseInstanceId}/sets/{setId}
     * @secure
     */
    routineServiceRemoveSetFromExerciseInstance: (
      routineId: string,
      exerciseInstanceId: string,
      setId: string,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/${exerciseInstanceId}/sets/${setId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceUpdateSetInExerciseInstance
     * @summary Метод для обновления сета в упражнении
     * @request PUT:/v1/routines/{routineId}/exercise_instances/{exerciseInstanceId}/sets/{setId}
     * @secure
     */
    routineServiceUpdateSetInExerciseInstance: (
      routineId: string,
      exerciseInstanceId: string,
      setId: string,
      body: RoutineServiceUpdateSetInExerciseInstanceBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutSetResponse, RpcStatus>({
        path: `/v1/routines/${routineId}/exercise_instances/${exerciseInstanceId}/sets/${setId}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags RoutineService
     * @name RoutineServiceAddExerciseToRoutine
     * @summary Добавление упражнения в рутину
     * @request POST:/v1/routines/{routineId}/exercises
     * @secure
     */
    routineServiceAddExerciseToRoutine: (
      routineId: string,
      body: RoutineServiceAddExerciseToRoutineBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutRoutineInstanceResponse, RpcStatus>({
        path: `/v1/routines/${routineId}/exercises`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceUpdateUser
     * @summary Метод для обновления данных пользователя
     * @request PUT:/v1/users
     * @secure
     */
    userServiceUpdateUser: (body: WorkoutUpdateUserRequest, params: RequestParams = {}) =>
      this.request<WorkoutUserResponse, RpcStatus>({
        path: `/v1/users`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceGetMe
     * @summary Метод для получения текущего пользователя
     * @request GET:/v1/users/me
     * @secure
     */
    userServiceGetMe: (params: RequestParams = {}) =>
      this.request<WorkoutUserResponse, RpcStatus>({
        path: `/v1/users/me`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceGetWorkoutGenerationSettings
     * @summary Метод для получения настроек генерации тренировок
     * @request GET:/v1/users/workout_generation_settings
     * @secure
     */
    userServiceGetWorkoutGenerationSettings: (params: RequestParams = {}) =>
      this.request<WorkoutWorkoutGenerationSettingsResponse, RpcStatus>({
        path: `/v1/users/workout_generation_settings`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceUpdateWorkoutGenerationSettings
     * @summary Метод для обновления настроек генерации тренировок
     * @request PUT:/v1/users/workout_generation_settings
     * @secure
     */
    userServiceUpdateWorkoutGenerationSettings: (
      body: WorkoutUpdateWorkoutGenerationSettingsRequest,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/users/workout_generation_settings`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags UserService
     * @name UserServiceGetUser
     * @summary Метод для получения пользователя по ID
     * @request GET:/v1/users/{userId}
     * @secure
     */
    userServiceGetUser: (userId: string, params: RequestParams = {}) =>
      this.request<WorkoutUserResponse, RpcStatus>({
        path: `/v1/users/${userId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGetWorkouts
     * @summary Метод для получения списка всех тренировок
     * @request GET:/v1/workouts
     * @secure
     */
    workoutServiceGetWorkouts: (
      query?: {
        /** @format int32 */
        offset?: number;
        /** @format int32 */
        limit?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<WorkoutGetWorkoutsResponse, RpcStatus>({
        path: `/v1/workouts`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceStartWorkout
     * @summary Метод для начала новой тренировки
     * @request POST:/v1/workouts
     * @secure
     */
    workoutServiceStartWorkout: (body: WorkoutStartWorkoutRequest, params: RequestParams = {}) =>
      this.request<WorkoutWorkoutResponse, RpcStatus>({
        path: `/v1/workouts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGetActiveWorkouts
     * @summary Метод для получения списка активных тренировок
     * @request GET:/v1/workouts/active
     * @secure
     */
    workoutServiceGetActiveWorkouts: (params: RequestParams = {}) =>
      this.request<WorkoutWorkoutsListResponse, RpcStatus>({
        path: `/v1/workouts/active`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGetWorkout
     * @summary Метод для получения состояния тренировки
     * @request GET:/v1/workouts/{workoutId}
     * @secure
     */
    workoutServiceGetWorkout: (workoutId: string, params: RequestParams = {}) =>
      this.request<WorkoutGetWorkoutResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceDeleteWorkout
     * @summary Удалить тренировку
     * @request DELETE:/v1/workouts/{workoutId}
     * @secure
     */
    workoutServiceDeleteWorkout: (workoutId: string, params: RequestParams = {}) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceAddCommentToWorkout
     * @summary Метод для добавления комментария к тренировке
     * @request POST:/v1/workouts/{workoutId}/comment
     * @secure
     */
    workoutServiceAddCommentToWorkout: (
      workoutId: string,
      body: WorkoutServiceAddCommentToWorkoutBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutWorkoutResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/comment`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceCompleteWorkout
     * @summary Метод для завершения тренировки
     * @request POST:/v1/workouts/{workoutId}/complete
     * @secure
     */
    workoutServiceCompleteWorkout: (
      workoutId: string,
      body: WorkoutServiceCompleteWorkoutBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}/complete`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGenerateWorkout
     * @summary Запустить (или перезапустить после ошибки) генерацию тренировки
     * @request POST:/v1/workouts/{workoutId}/generate
     * @secure
     */
    workoutServiceGenerateWorkout: (
      workoutId: string,
      body: WorkoutServiceGenerateWorkoutBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutWorkoutResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/generate`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceLogExercise
     * @summary Метод для создания записи о выполнении упражнения
     * @request POST:/v1/workouts/{workoutId}/log/exercise
     * @secure
     */
    workoutServiceLogExercise: (workoutId: string, body: WorkoutServiceLogExerciseBody, params: RequestParams = {}) =>
      this.request<WorkoutExerciseLog, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGetExerciseLogDetails
     * @summary Метод для получения записи о выполнении упражнения
     * @request GET:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}
     * @secure
     */
    workoutServiceGetExerciseLogDetails: (workoutId: string, exerciseLogId: string, params: RequestParams = {}) =>
      this.request<WorkoutExerciseLogResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceDeleteExerciseLog
     * @summary Удалить запись о выполнении упражнения
     * @request DELETE:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}
     * @secure
     */
    workoutServiceDeleteExerciseLog: (workoutId: string, exerciseLogId: string, params: RequestParams = {}) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceAddNotesToExerciseLog
     * @summary Добавление заметки к выполнению упражнения
     * @request POST:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/notes
     * @secure
     */
    workoutServiceAddNotesToExerciseLog: (
      workoutId: string,
      exerciseLogId: string,
      body: WorkoutServiceAddNotesToExerciseLogBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/notes`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceAddPowerRatingToExerciseLog
     * @summary Добавить оценку усилий при выполнении упражнения
     * @request POST:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/power_rating
     * @secure
     */
    workoutServiceAddPowerRatingToExerciseLog: (
      workoutId: string,
      exerciseLogId: string,
      body: WorkoutServiceAddPowerRatingToExerciseLogBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/power_rating`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceLogSet
     * @summary Метод для создания записи о выполнении подхода
     * @request POST:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/set
     * @secure
     */
    workoutServiceLogSet: (
      workoutId: string,
      exerciseLogId: string,
      body: WorkoutServiceLogSetBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutSetLogResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/set`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceDeleteSetLog
     * @summary Удалить запись о выполнении подхода
     * @request DELETE:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/set/{setId}
     * @secure
     */
    workoutServiceDeleteSetLog: (workoutId: string, exerciseLogId: string, setId: string, params: RequestParams = {}) =>
      this.request<WorkoutServiceCompleteWorkoutBody, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/set/${setId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceUpdateSetLog
     * @summary Метод для изменения записи о выполнении подхода
     * @request PUT:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/set/{setId}
     * @secure
     */
    workoutServiceUpdateSetLog: (
      workoutId: string,
      exerciseLogId: string,
      setId: string,
      body: WorkoutServiceUpdateSetLogBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutSetLogResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/set/${setId}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceUpdateExerciseLogWeightUnit
     * @summary Изменить единицу измерения веса для ExerciseLog
     * @request PATCH:/v1/workouts/{workoutId}/log/exercise/{exerciseLogId}/weight_unit
     * @secure
     */
    workoutServiceUpdateExerciseLogWeightUnit: (
      workoutId: string,
      exerciseLogId: string,
      body: WorkoutServiceUpdateExerciseLogWeightUnitBody,
      params: RequestParams = {},
    ) =>
      this.request<WorkoutExerciseLogResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/log/exercise/${exerciseLogId}/weight_unit`,
        method: "PATCH",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceRateWorkout
     * @summary Метод для установки оценки тренировки
     * @request POST:/v1/workouts/{workoutId}/rate
     * @secure
     */
    workoutServiceRateWorkout: (workoutId: string, body: WorkoutServiceRateWorkoutBody, params: RequestParams = {}) =>
      this.request<WorkoutWorkoutResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/rate`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags WorkoutService
     * @name WorkoutServiceGetWorkoutReport
     * @summary Метод для получения отчета о тренировке
     * @request GET:/v1/workouts/{workoutId}/report
     * @secure
     */
    workoutServiceGetWorkoutReport: (workoutId: string, params: RequestParams = {}) =>
      this.request<WorkoutWorkoutReportResponse, RpcStatus>({
        path: `/v1/workouts/${workoutId}/report`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
